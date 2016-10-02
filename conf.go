package main

type Config struct {
	Dropbox struct {
		ApiKey      string `env:"DROPBOX_API_KEY,required"`
		ApiSecret   string `env:"DROPBOX_API_SECRET,required"`
		RedirectURL string `env:"DROPBOX_REDIRECT_URL,default=https://dropbox.rsampaio.info/oauth"`
	}

	Google struct {
	}
}
