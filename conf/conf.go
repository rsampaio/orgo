package conf

// Config struct exposes configuration keys read from environment variables
type Config struct {
	// Secret to encrypt cookies
	HTTPCookieSecret string `env:"HTTP_COOKIE_SECRET,default=secretkey123"`

	// Dropbox parameters
	Dropbox struct {
		APIKey      string `env:"DROPBOX_API_KEY,required"`
		APISecret   string `env:"DROPBOX_API_SECRET,required"`
		RedirectURL string `env:"DROPBOX_REDIRECT_URL,default=https://orgo.rsampaio.info/dropbox/oauth"`
	}

	// Google parameters
	Google struct {
		APIKey      string `env:"GOOGLE_API_KEY,required"`
		APISecret   string `env:"GOOGLE_API_SECRET,required"`
		RedirectURL string `env:"GOOGLE_REDIRECT_URL,default=postmessage"`
	}
}
