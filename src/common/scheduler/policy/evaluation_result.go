package policy

//EvaluationResult is defined to carry the policy evaluated result.
//
//Filed 'Result' is optional.
//Filed 'Error' is optional
//
type EvaluationResult struct {
	//Policy is successfully evaluated and the related information can
	//be contained in Result if have.
	Result interface{}

	//Policy is failed to evaluated.
	Error error
}
