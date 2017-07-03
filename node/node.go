package node

type Options struct {
	Binding           string
	UseTLS            bool
	CertFilePath      string
	KeyFilePath       string
	TrustedCAFilePath string
	LogFile           bool
	LogPath           string
	Syslog            bool
}

func Serve(options Options) error {
	return nil
}
