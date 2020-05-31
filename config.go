package sendemail

type ServerConfig struct {
	Host   string
	Port   int
	Title  string
	DryRun bool
}

type SmtpConfig struct {
	FromDomain string
	Host       string
	Port       int
	Username   string
	Password   string
	Bcc        []string
}

type Config struct {
	Server ServerConfig
	Sender map[string]SmtpConfig
}
