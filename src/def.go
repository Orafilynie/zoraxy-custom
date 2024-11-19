package main

/*
	Type and flag definations

	This file contains all the type and flag definations
	Author: tobychui
*/

import (
	"embed"
	"flag"
	"net/http"
	"time"

	"imuslab.com/zoraxy/mod/access"
	"imuslab.com/zoraxy/mod/acme"
	"imuslab.com/zoraxy/mod/auth"
	"imuslab.com/zoraxy/mod/auth/sso"
	"imuslab.com/zoraxy/mod/database"
	"imuslab.com/zoraxy/mod/dockerux"
	"imuslab.com/zoraxy/mod/dynamicproxy/loadbalance"
	"imuslab.com/zoraxy/mod/dynamicproxy/redirection"
	"imuslab.com/zoraxy/mod/email"
	"imuslab.com/zoraxy/mod/forwardproxy"
	"imuslab.com/zoraxy/mod/ganserv"
	"imuslab.com/zoraxy/mod/geodb"
	"imuslab.com/zoraxy/mod/info/logger"
	"imuslab.com/zoraxy/mod/info/logviewer"
	"imuslab.com/zoraxy/mod/mdns"
	"imuslab.com/zoraxy/mod/netstat"
	"imuslab.com/zoraxy/mod/pathrule"
	"imuslab.com/zoraxy/mod/sshprox"
	"imuslab.com/zoraxy/mod/statistic"
	"imuslab.com/zoraxy/mod/statistic/analytic"
	"imuslab.com/zoraxy/mod/streamproxy"
	"imuslab.com/zoraxy/mod/tlscert"
	"imuslab.com/zoraxy/mod/uptime"
	"imuslab.com/zoraxy/mod/webserv"
)

const (
	/* Build Constants */
	SYSTEM_NAME       = "Zoraxy"
	SYSTEM_VERSION    = "3.1.4"
	DEVELOPMENT_BUILD = true /* Development: Set to false to use embedded web fs */

	/* System Constants */
	DATABASE_PATH              = "sys.db"
	TMP_FOLDER                 = "./tmp"
	WEBSERV_DEFAULT_PORT       = 5487
	MDNS_HOSTNAME_PREFIX       = "zoraxy_" /* Follow by node UUID */
	MDNS_IDENTIFY_DEVICE_TYPE  = "Network Gateway"
	MDNS_IDENTIFY_DOMAIN       = "zoraxy.aroz.org"
	MDNS_IDENTIFY_VENDOR       = "imuslab.com"
	MDNS_SCAN_TIMEOUT          = 30 /* Seconds */
	MDNS_SCAN_UPDATE_INTERVAL  = 15 /* Minutes */
	ACME_AUTORENEW_CONFIG_PATH = "./conf/acme_conf.json"
	CSRF_COOKIENAME            = "zoraxy_csrf"
	LOG_PREFIX                 = "zr"
	LOG_FOLDER                 = "./log"
	LOG_EXTENSION              = ".log"

	/* Configuration Folder Storage Path Constants */
	CONF_HTTP_PROXY   = "./conf/proxy"
	CONF_STREAM_PROXY = "./conf/streamproxy"
	CONF_CERT_STORE   = "./conf/certs"
	CONF_REDIRECTION  = "./conf/redirect"
	CONF_ACCESS_RULE  = "./conf/access"
	CONF_PATH_RULE    = "./conf/rules/pathrules"
)

/* System Startup Flags */
var webUIPort = flag.String("port", ":8000", "Management web interface listening port")
var noauth = flag.Bool("noauth", false, "Disable authentication for management interface")
var showver = flag.Bool("version", false, "Show version of this server")
var allowSshLoopback = flag.Bool("sshlb", false, "Allow loopback web ssh connection (DANGER)")
var allowMdnsScanning = flag.Bool("mdns", true, "Enable mDNS scanner and transponder")
var mdnsName = flag.String("mdnsname", "", "mDNS name, leave empty to use default (zoraxy_{node-uuid}.local)")
var ztAuthToken = flag.String("ztauth", "", "ZeroTier authtoken for the local node")
var ztAPIPort = flag.Int("ztport", 9993, "ZeroTier controller API port")
var runningInDocker = flag.Bool("docker", false, "Run Zoraxy in docker compatibility mode")
var acmeAutoRenewInterval = flag.Int("autorenew", 86400, "ACME auto TLS/SSL certificate renew check interval (seconds)")
var acmeCertAutoRenewDays = flag.Int("earlyrenew", 30, "Number of days to early renew a soon expiring certificate (days)")
var enableHighSpeedGeoIPLookup = flag.Bool("fastgeoip", false, "Enable high speed geoip lookup, require 1GB extra memory (Not recommend for low end devices)")
var staticWebServerRoot = flag.String("webroot", "./www", "Static web server root folder. Only allow chnage in start paramters")
var allowWebFileManager = flag.Bool("webfm", true, "Enable web file manager for static web server root folder")
var enableAutoUpdate = flag.Bool("cfgupgrade", true, "Enable auto config upgrade if breaking change is detected")

/* Global Variables and Handlers */
var (
	nodeUUID    = "generic" //System uuid, in uuidv4 format, load from database on startup
	bootTime    = time.Now().Unix()
	requireAuth = true /* Require authentication for webmin panel */

	/*
		Binary Embedding File System
	*/
	//go:embed web/*
	webres embed.FS

	/*
		Handler Modules
	*/
	sysdb          *database.Database              //System database
	authAgent      *auth.AuthAgent                 //Authentication agent
	tlsCertManager *tlscert.Manager                //TLS / SSL management
	redirectTable  *redirection.RuleTable          //Handle special redirection rule sets
	webminPanelMux *http.ServeMux                  //Server mux for handling webmin panel APIs
	csrfMiddleware func(http.Handler) http.Handler //CSRF protection middleware

	pathRuleHandler    *pathrule.Handler         //Handle specific path blocking or custom headers
	geodbStore         *geodb.Store              //GeoIP database, for resolving IP into country code
	accessController   *access.Controller        //Access controller, handle black list and white list
	netstatBuffers     *netstat.NetStatBuffers   //Realtime graph buffers
	statisticCollector *statistic.Collector      //Collecting statistic from visitors
	uptimeMonitor      *uptime.Monitor           //Uptime monitor service worker
	mdnsScanner        *mdns.MDNSHost            //mDNS discovery services
	ganManager         *ganserv.NetworkManager   //Global Area Network Manager
	webSshManager      *sshprox.Manager          //Web SSH connection service
	streamProxyManager *streamproxy.Manager      //Stream Proxy Manager for TCP / UDP forwarding
	acmeHandler        *acme.ACMEHandler         //Handler for ACME Certificate renew
	acmeAutoRenewer    *acme.AutoRenewer         //Handler for ACME auto renew ticking
	staticWebServer    *webserv.WebServer        //Static web server for hosting simple stuffs
	forwardProxy       *forwardproxy.Handler     //HTTP Forward proxy, basically VPN for web browser
	loadBalancer       *loadbalance.RouteManager //Global scope loadbalancer, store the state of the lb routing
	ssoHandler         *sso.SSOHandler           //Single Sign On handler

	//Helper modules
	EmailSender       *email.Sender         //Email sender that handle email sending
	AnalyticLoader    *analytic.DataLoader  //Data loader for Zoraxy Analytic
	DockerUXOptimizer *dockerux.UXOptimizer //Docker user experience optimizer, community contribution only
	SystemWideLogger  *logger.Logger        //Logger for Zoraxy
	LogViewer         *logviewer.Viewer
)
