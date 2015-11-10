package main

type Repositorer interface {
	Open()
	Close()
	Lock()
	Unlock()
}

func NewRepository() *Repository {
	return nil
}

func CloneRepository() *Repository {
	return nil
}

type Repository struct {
	Jid      string // name@domain.tld
	Password string // sha-x representation
	Folder   string // filesystem foldername repo is in
	Uid      string

	ConfigPath string

	// Crypto
	PublicKey  string
	PrivateKey string
	AesKey     string
	OtrKey     string
}

func main() {
}
