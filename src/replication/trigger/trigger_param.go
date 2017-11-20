package trigger

//BasicParam contains the general parameters for all triggers
type BasicParam struct {
	//ID of the related policy
	PolicyID int64

	//Whether delete remote replicated images if local ones are deleted
	OnDeletion bool
}

//Parameter defines operation of doing initialization from parameter json text
type Parameter interface {
	//Decode parameter with json style to the owner struct
	//If failed, an error will be returned
	Parse(param string) error
}
