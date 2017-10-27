package source

import (
	"github.com/vmware/harbor/src/replication/models"
)

//Convertor is designed to covert the format of output from upstream filter to the input format
//required by the downstream filter if needed.
//Each convertor covers only one specified conversion process between the two filters.
//E.g:
//If project filter connects to repository filter, then one convertor should be defined for this connection;
//If project filter connects to tag filter, then another one should be defined. The above one can not be reused.
type Convertor interface {
	//Accept the items from upstream filter as input and then covert them to the required format and returned.
	Convert(itemsOfUpstream []models.FilterItem) (itemsOfDownstream []models.FilterItem)
}
