package main

type Config struct {
	HttpCookieSecret string `env:"HTTP_COOKIE_SECRET,default=secretkey123"`
	Dropbox          struct {
		ApiKey      string `env:"DROPBOX_API_KEY,required"`
		ApiSecret   string `env:"DROPBOX_API_SECRET,required"`
		RedirectURL string `env:"DROPBOX_REDIRECT_URL,default=https://orgo.rsampaio.info/dropbox/oauth"`
	}

	Google struct {
		ApiKey      string `env:"GOOGLE_API_KEY,required"`
		ApiSecret   string `env:"GOOGLE_API_SECRET,required"`
		RedirectURL string `env:"GOOGLE_REDIRECT_URL,default=postmessage"`
	}
}
