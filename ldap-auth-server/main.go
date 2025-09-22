package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"crypto/tls"
	"crypto/x509"

	"github.com/gin-gonic/gin"
	ldap "github.com/go-ldap/ldap/v3"
	log "github.com/sirupsen/logrus"
)

var configFile = flag.String("config", "/data/options.json", "The path to the server's config file")
var errCouldNotParseCAFile = errors.New("failed to parse certificates")

// LDAPOptions holds the opotions for an `LDAPAuthenticator`.
type LDAPOptions struct {
	ServerURL            string
	BindDNTemplate       string
	SearchBaseDN         string
	SearchFilterTemplate string
	TLSConfig            *tls.Config
}

// LDAPUser holds metadata about a user from an LDAP server.
type LDAPUser struct {
	DisplayName string
}

// LDAPAuthenticator is a client for authenticating users against an LDAP server.
type LDAPAuthenticator struct {
	Options LDAPOptions
}

// Authenticate authenticates a user given their username and password.
//
// If authentication is successful, it also returns metadata about the
// authenticated user.
func (a *LDAPAuthenticator) Authenticate(username, password string) (bool, LDAPUser, error) {
	if password == "" {
		return false, LDAPUser{}, nil
	}

	var dialOpts []ldap.DialOpt
	if a.Options.TLSConfig != nil {
		dialOpts = append(dialOpts, ldap.DialWithTLSConfig(a.Options.TLSConfig))
	}
	conn, err := ldap.DialURL(a.Options.ServerURL, dialOpts...)
	if err != nil {
		return false, LDAPUser{}, fmt.Errorf("could not dial server at %q: %w", a.Options.ServerURL, err)
	}
	defer conn.Close()

	bindDN := fmt.Sprintf(a.Options.BindDNTemplate, username)
	log.Debugf("Binding to DN: %q", bindDN)
	if err := conn.Bind(bindDN, password); err != nil {
		log.Debugf("Could not bind to user %q: %q", bindDN, err)
		return false, LDAPUser{}, nil
	}

	var searchBaseDN string
	if a.Options.SearchBaseDN == "" {
		res, err := conn.WhoAmI( /* controls= */ nil)
		if err != nil {
			return false, LDAPUser{}, fmt.Errorf("WhoAmiI failed: %w", err)
		}
		if res.AuthzID != "" {
			searchBaseDN = strings.TrimPrefix(res.AuthzID, "dn:")
		} else {
			searchBaseDN = bindDN
		}
	} else {
		searchBaseDN = a.Options.SearchBaseDN
	}

	var searchFilter string
	if a.Options.SearchFilterTemplate == "" {
		// Meant to be a filter that always matches.
		searchFilter = "(objectClass=*)"
	} else if strings.Contains(a.Options.SearchFilterTemplate, "%s") {
		searchFilter = fmt.Sprintf(a.Options.SearchFilterTemplate, username)
	} else {
		searchFilter = a.Options.SearchFilterTemplate
	}

	log.Debugf("Searching with baseDN %s and filter: %q", searchBaseDN, searchFilter)
	res, err := conn.Search(ldap.NewSearchRequest(
		searchBaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		/* SizeLimit= */ 2,
		/* TimeLimit= */ 1,
		/* TypesOnly= */ false,
		searchFilter,
		/* Attributes= */ nil,
		/* Controls= */ nil,
	))
	if err != nil {
		log.Warnf("Search failed: %v", err)
		return false, LDAPUser{}, nil
	}

	if len(res.Entries) == 0 {
		log.Debugf("No matches found for filter: %q", searchFilter)
		return false, LDAPUser{}, nil
	}
	if len(res.Entries) > 1 {
		log.Debugf("More than one (%d) matches found for filter: %q", len(res.Entries), searchFilter)
		return false, LDAPUser{}, nil
	}

	entry := res.Entries[0]
	if log.IsLevelEnabled(log.DebugLevel) {
		var entryStr strings.Builder
		fmt.Fprintf(&entryStr, "DN: %q\n", entry.DN)
		for _, attr := range entry.Attributes {
			fmt.Fprintf(&entryStr, "%s: %q\n", attr.Name, attr.Values)
		}
		log.Debugf("User entry:\n%v", entryStr.String())
	}
	return true, LDAPUser{
		DisplayName: entry.GetAttributeValue("cn"),
	}, nil
}

var ErrUserAuthnFailed = errors.New("could not authenticate user")

type ServerOptions struct {
	LDAPOptions LDAPOptions
}

type HASSAuthenticateRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ErrorResponse struct {
	StatusCode  int
	PublicError string
	Error       error
}

type Server struct {
	options       ServerOptions
	authenticator LDAPAuthenticator
	router        *gin.Engine
}

// NewServer creates a new Server.
func NewServer(options ServerOptions) *Server {
	server := &Server{
		options: options,
		authenticator: LDAPAuthenticator{
			Options: options.LDAPOptions,
		},
		router: gin.Default(),
	}
	if err := server.router.SetTrustedProxies(nil); err != nil {
		panic("Not possible")
	}
	server.router.POST("/hass_authenticate", server.hassAuthenticate)
	return server
}

func (s *Server) writeError(c *gin.Context, err ErrorResponse) {
	if err.StatusCode >= 400 && err.StatusCode < 500 {
		log.Infof("%s", err.Error.Error())
	} else {
		log.Warnf("%s", err.Error.Error())
	}

	if gin.Mode() == gin.DebugMode {
		c.JSON(err.StatusCode, gin.H{"error": err.Error.Error()})
	} else {
		c.JSON(err.StatusCode, gin.H{"error": err.PublicError})
	}
}

func (s *Server) hassAuthenticate(c *gin.Context) {
	var req HASSAuthenticateRequest
	if err := c.BindJSON(&req); err != nil {
		s.writeError(c, ErrorResponse{
			StatusCode:  http.StatusBadRequest,
			PublicError: err.Error(),
			Error:       err,
		})
		return
	}

	ok, user, err := s.authenticator.Authenticate(req.Username, req.Password)
	if err != nil {
		s.writeError(c, ErrorResponse{
			StatusCode:  http.StatusBadRequest,
			PublicError: fmt.Sprintf("Could not authenticate user %q", req.Username),
			Error:       fmt.Errorf("when authenticating: %w", err),
		})
		return
	}
	if !ok {
		s.writeError(c, ErrorResponse{
			StatusCode:  http.StatusBadRequest,
			PublicError: fmt.Sprintf("Could not authenticate user: %q", req.Username),
			Error:       fmt.Errorf("when authenticating %q: %w", req.Username, ErrUserAuthnFailed),
		})
		return
	}
	if user.DisplayName != "" {
		safeDN := strings.ReplaceAll(user.DisplayName, "\n", "")
		fmt.Fprintf(c.Writer, "name = %v", safeDN)
	}
}

// Serve starts the server listening to requests.
func (s *Server) Serve() error {
	addr := ":80"
	log.Infof("Starting server at: %q", addr)
	return s.router.Run(addr)
}

// AddOnConfig is the struct representing the add-on's configuration.
type AddOnConfig struct {
	LDAPServerURL               string `json:"ldap_server_url"`
	BindDNTemplate              string `json:"bind_dn_template"`
	SearchBaseDN                string `json:"search_base_dn"`
	SearchFilterTemplate        string `json:"search_filter_template"`
	ServerRootCAsFile           string `json:"server_root_cas_file"`
	DisableServerCertValidation bool   `json:"disable_server_cert_validation"`
	DebugMode                   bool   `json:"debug_mode"`
}

// parseAddOnConfig parses the `AddOnConfig` from a JSON file.
func parseAddOnConfig(configFile string) (AddOnConfig, error) {
	jsonBytes, err := os.ReadFile(configFile)
	if err != nil {
		return AddOnConfig{}, fmt.Errorf("could not read config file: %q: %w", configFile, err)
	}

	var config AddOnConfig
	err = json.Unmarshal(jsonBytes, &config)
	if err != nil {
		return AddOnConfig{}, fmt.Errorf("could not parse config file: %q: %w", configFile, err)
	}

	if (config.SearchBaseDN == "") != (config.SearchFilterTemplate == "") {
		return AddOnConfig{}, fmt.Errorf("search_base_dn (%q) and search_filter_template (%q) must both be set or both be unset: %w", config.SearchBaseDN, config.SearchFilterTemplate, ErrInvalidServerOptions)
	}

	return config, nil
}

var ErrInvalidServerOptions = errors.New("invalid server options")

func toServerOptions(config AddOnConfig) (ServerOptions, error) {
	log.Infof("Loaded config: %+v", config)

	var tlsConfig *tls.Config
	if strings.HasPrefix(strings.ToLower(config.LDAPServerURL), "ldaps://") {
		if config.DisableServerCertValidation {
			log.Warn("Disabling server certificate validation")
		}

		tlsConfig = &tls.Config{
			InsecureSkipVerify: config.DisableServerCertValidation,
		}
		// Load custom root CAs if provided
		if config.ServerRootCAsFile != "" {
			rootCAFile := fmt.Sprintf("/config/%s", config.ServerRootCAsFile)
			log.Infof("Loading custom root CAs from file %q", rootCAFile)
			caData, err := os.ReadFile(rootCAFile)
			if err != nil {
				return ServerOptions{}, fmt.Errorf("could not read server_root_cas_file %q: %w", config.ServerRootCAsFile, err)
			}
			certPool := x509.NewCertPool()
			if !certPool.AppendCertsFromPEM(caData) {
				return ServerOptions{}, fmt.Errorf("from %q: %w", config.ServerRootCAsFile, errCouldNotParseCAFile)
			}
			tlsConfig.RootCAs = certPool
		}
	}

	return ServerOptions{
		LDAPOptions: LDAPOptions{
			ServerURL:            config.LDAPServerURL,
			BindDNTemplate:       config.BindDNTemplate,
			SearchBaseDN:         config.SearchBaseDN,
			SearchFilterTemplate: config.SearchFilterTemplate,
			TLSConfig:            tlsConfig,
		},
	}, nil
}

func init() {
	log.SetReportCaller(true)
	log.SetLevel(log.InfoLevel)
}

func main() {
	flag.Parse()
	config, err := parseAddOnConfig(*configFile)
	if err != nil {
		log.Fatalf("Invalid config: %v", err)
	}
	if config.DebugMode {
		gin.SetMode(gin.DebugMode)
		log.SetLevel(log.DebugLevel)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	options, err := toServerOptions(config)
	if err != nil {
		log.Fatalf("Invalid config: %v", err)
	}
	server := NewServer(options)
	if err := server.Serve(); err != nil {
		log.Fatalf("Error while serving: %v", err)
	}
	log.Info("Server terminated.")
}
