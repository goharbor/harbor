package trustpinning_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"text/template"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/notary"
	"github.com/docker/notary/cryptoservice"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/trustpinning"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/docker/notary/tuf/testutils"
	"github.com/docker/notary/tuf/utils"
	"github.com/stretchr/testify/require"
)

type SignedRSARootTemplate struct {
	RootPem string
}

var passphraseRetriever = func(string, string, bool, int) (string, bool, error) { return "passphrase", false, nil }

const validPEMEncodedRSARoot = `LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUZLekNDQXhXZ0F3SUJBZ0lRUnlwOVFxY0pmZDNheXFkaml6OHhJREFMQmdrcWhraUc5dzBCQVFzd09ERWEKTUJnR0ExVUVDaE1SWkc5amEyVnlMbU52YlM5dWIzUmhjbmt4R2pBWUJnTlZCQU1URVdSdlkydGxjaTVqYjIwdgpibTkwWVhKNU1CNFhEVEUxTURjeE56QTJNelF5TTFvWERURTNNRGN4TmpBMk16UXlNMW93T0RFYU1CZ0dBMVVFCkNoTVJaRzlqYTJWeUxtTnZiUzl1YjNSaGNua3hHakFZQmdOVkJBTVRFV1J2WTJ0bGNpNWpiMjB2Ym05MFlYSjUKTUlJQ0lqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FnOEFNSUlDQ2dLQ0FnRUFvUWZmcnpzWW5zSDh2R2Y0Smg1NQpDajV3cmpVR3pEL3NIa2FGSHB0ako2VG9KR0p2NXlNQVB4enlJbnU1c0lvR0xKYXBuWVZCb0FVMFlnSTlxbEFjCllBNlN4YVN3Z202cnB2bW5sOFFuMHFjNmdlcjNpbnBHYVVKeWxXSHVQd1drdmNpbVFBcUhaeDJkUXRMN2c2a3AKcm1LZVRXcFdvV0x3M0pvQVVaVVZoWk1kNmEyMlpML0R2QXcrSHJvZ2J6NFhleWFoRmI5SUg0MDJ6UHhONnZnYQpKRUZURjBKaTFqdE5nME1vNHBiOVNIc01zaXcrTFpLN1NmZkhWS1B4dmQyMW0vYmlObXdzZ0V4QTNVOE9PRzhwCnV5Z2ZhY3lzNWM4K1pyWCtaRkcvY3Z3S3owazYvUWZKVTQwczZNaFh3NUMyV3R0ZFZtc0c5LzdyR0ZZakhvSUoKd2VEeXhnV2s3dnhLelJKSS91bjdjYWdESWFRc0tySlFjQ0hJR0ZSbHBJUjVUd1g3dmwzUjdjUm5jckRSTVZ2YwpWU0VHMmVzeGJ3N2p0eklwL3lwblZSeGNPbnk3SXlweWpLcVZlcVo2SGd4WnRUQlZyRjFPL2FIbzJrdmx3eVJTCkF1czRrdmg2ejMranpUbTlFemZYaVBRelk5QkVrNWdPTHhoVzlyYzZVaGxTK3BlNWxrYU4vSHlxeS9sUHVxODkKZk1yMnJyN2xmNVdGZEZuemU2V05ZTUFhVzdkTkE0TkUwZHlENTM0MjhaTFh4TlZQTDRXVTY2R2FjNmx5blE4bApyNXRQc1lJRlh6aDZGVmFSS0dRVXRXMWh6OWVjTzZZMjdSaDJKc3lpSXhnVXFrMm9veEU2OXVONDJ0K2R0cUtDCjFzOEcvN1Z0WThHREFMRkxZVG56THZzQ0F3RUFBYU0xTURNd0RnWURWUjBQQVFIL0JBUURBZ0NnTUJNR0ExVWQKSlFRTU1Bb0dDQ3NHQVFVRkJ3TURNQXdHQTFVZEV3RUIvd1FDTUFBd0N3WUpLb1pJaHZjTkFRRUxBNElDQVFCTQpPbGwzRy9YQno4aWRpTmROSkRXVWgrNXczb2ptd2FuclRCZENkcUVrMVdlbmFSNkR0Y2ZsSng2WjNmL213VjRvCmIxc2tPQVgxeVg1UkNhaEpIVU14TWljei9RMzhwT1ZlbEdQclduYzNUSkIrVktqR3lIWGxRRFZrWkZiKzQrZWYKd3RqN0huZ1hoSEZGRFNnam0zRWRNbmR2Z0RRN1NRYjRza09uQ05TOWl5WDdlWHhoRkJDWm1aTCtIQUxLQmoyQgp5aFY0SWNCRHFtcDUwNHQxNHJ4OS9KdnR5MGRHN2ZZN0k1MWdFUXBtNFMwMkpNTDV4dlRtMXhmYm9XSWhaT0RJCnN3RUFPK2VrQm9GSGJTMVE5S01QaklBdzNUckNISDh4OFhacTV6c1l0QUMxeVpIZENLYTI2YVdkeTU2QTllSGoKTzFWeHp3bWJOeVhSZW5WdUJZUCswd3IzSFZLRkc0Sko0WlpwTlp6UVcvcHFFUGdoQ1RKSXZJdWVLNjUyQnlVYwovL3N2K25YZDVmMTlMZUVTOXBmMGwyNTNORGFGWlBiNmFlZ0tmcXVXaDhxbFFCbVVRMkd6YVRMYnRtTmQyOE02Clc3aUw3dGtLWmUxWm5CejlSS2d0UHJEampXR1pJbmpqY09VOEV0VDRTTHE3a0NWRG1QczVNRDh2YUFtOTZKc0UKam1MQzNVdS80azdIaURZWDBpMG1PV2tGalpRTWRWYXRjSUY1RlBTcHB3c1NiVzhRaWRuWHQ1NFV0d3RGREVQegpscGpzN3liZVFFNzFKWGNNWm5WSUs0YmpSWHNFRlBJOThScElsRWRlZGJTVWRZQW5jTE5KUlQ3SFpCTVBHU3daCjBQTkp1Z2xubHIzc3JWemRXMWR6MnhRamR2THd4eTZtTlVGNnJiUUJXQT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K`

const validCAPEMEncodeRSARoot = `LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tDQpNSUlHTXpDQ0JCdWdBd0lCQWdJQkFUQU5CZ2txaGtpRzl3MEJBUXNGQURCZk1Rc3dDUVlEVlFRR0V3SlZVekVMDQpNQWtHQTFVRUNBd0NRMEV4RmpBVUJnTlZCQWNNRFZOaGJpQkdjbUZ1WTJselkyOHhEekFOQmdOVkJBb01Ca1J2DQpZMnRsY2pFYU1CZ0dBMVVFQXd3UlRtOTBZWEo1SUZSbGMzUnBibWNnUTBFd0hoY05NVFV3TnpFMk1EUXlOVEF6DQpXaGNOTWpVd056RXpNRFF5TlRBeldqQmZNUm93R0FZRFZRUUREQkZPYjNSaGNua2dWR1Z6ZEdsdVp5QkRRVEVMDQpNQWtHQTFVRUJoTUNWVk14RmpBVUJnTlZCQWNNRFZOaGJpQkdjbUZ1WTJselkyOHhEekFOQmdOVkJBb01Ca1J2DQpZMnRsY2pFTE1Ba0dBMVVFQ0F3Q1EwRXdnZ0lpTUEwR0NTcUdTSWIzRFFFQkFRVUFBNElDRHdBd2dnSUtBb0lDDQpBUUN3VlZENHBLN3o3cFhQcEpiYVoxSGc1ZVJYSWNhWXRiRlBDbk4waXF5OUhzVkVHbkVuNUJQTlNFc3VQK20wDQo1TjBxVlY3REdiMVNqaWxvTFhEMXFERHZoWFdrK2dpUzlwcHFQSFBMVlBCNGJ2enNxd0RZcnRwYnFrWXZPMFlLDQowU0wza3hQWFVGZGxrRmZndTB4amxjem0yUGhXRzNKZDhhQXRzcEwvTCtWZlBBMTNKVWFXeFNMcHVpMUluOHJoDQpnQXlRVEs2UTRPZjZHYkpZVG5BSGI1OVVvTFhTekI1QWZxaVVxNkw3bkVZWUtvUGZsUGJSQUlXTC9VQm0wYytIDQpvY21zNzA2UFlwbVBTMlJRdjNpT0dtbm45aEVWcDNQNmpxN1dBZXZiQTRhWUd4NUVzYlZ0WUFCcUpCYkZXQXV3DQp3VEdSWW16bjBNajBlVE1nZTl6dFlCMi8yc3hkVGU2dWhtRmdwVVhuZ0RxSkk1TzlOM3pQZnZsRUltQ2t5M0hNDQpqSm9MN2c1c21xWDlvMVArRVNMaDBWWnpoaDdJRFB6UVRYcGNQSVMvNnowbDIyUUdrSy8xTjFQYUFEYVVIZExMDQp2U2F2M3kyQmFFbVB2ZjJma1pqOHlQNWVZZ2k3Q3c1T05oSExEWUhGY2w5Wm0veXdtZHhISkVUejluZmdYbnNXDQpITnhEcXJrQ1ZPNDZyL3U2clNyVXQ2aHIzb2RkSkc4czhKbzA2ZWFydzZYVTNNek0rM2dpd2tLMFNTTTN1UlBxDQo0QXNjUjFUditFMzFBdU9BbWpxWVFvVDI5Yk1JeG9TemVsamovWW5lZHdqVzQ1cFd5YzNKb0hhaWJEd3ZXOVVvDQpHU1pCVnk0aHJNL0ZhN1hDV3YxV2ZITlcxZ0R3YUxZd0RubDVqRm1SQnZjZnVRSURBUUFCbzRINU1JSDJNSUdSDQpCZ05WSFNNRWdZa3dnWWFBRkhVTTFVM0U0V3lMMW52RmQrZFBZOGY0TzJoWm9XT2tZVEJmTVFzd0NRWURWUVFHDQpFd0pWVXpFTE1Ba0dBMVVFQ0F3Q1EwRXhGakFVQmdOVkJBY01EVk5oYmlCR2NtRnVZMmx6WTI4eER6QU5CZ05WDQpCQW9NQmtSdlkydGxjakVhTUJnR0ExVUVBd3dSVG05MFlYSjVJRlJsYzNScGJtY2dRMEdDQ1FEQ2VETGJlbUlUDQpTekFTQmdOVkhSTUJBZjhFQ0RBR0FRSC9BZ0VBTUIwR0ExVWRKUVFXTUJRR0NDc0dBUVVGQndNQ0JnZ3JCZ0VGDQpCUWNEQVRBT0JnTlZIUThCQWY4RUJBTUNBVVl3SFFZRFZSME9CQllFRkhlNDhoY0JjQXAwYlVWbFR4WGVSQTRvDQpFMTZwTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElDQVFBV1V0QVBkVUZwd1JxK04xU3pHVWVqU2lrZU1HeVBac2NaDQpKQlVDbWhab0Z1ZmdYR2JMTzVPcGNSTGFWM1hkYTB0LzVQdGRHTVNFemN6ZW9aSFdrbkR0dys3OU9CaXR0UFBqDQpTaDFvRkR1UG8zNVI3ZVA2MjRsVUNjaC9JblpDcGhUYUx4OW9ETEdjYUszYWlsUTl3akJkS2RsQmw4S05LSVpwDQphMTNhUDVyblNtMkp2YSt0WHkveWkzQlNkczNkR0Q4SVRLWnlJLzZBRkh4R3ZPYnJESUJwbzRGRi96Y1dYVkRqDQpwYU9teHBsUnRNNEhpdG0rc1hHdmZxSmU0eDVEdU9YT25QclQzZEh2UlQ2dlNaVW9Lb2J4TXFtUlRPY3JPSVBhDQpFZU1wT29ic2hPUnVSbnRNRFl2dmdPM0Q2cDZpY2lEVzJWcDlONnJkTWRmT1dFUU44SlZXdkI3SXhSSGs5cUtKDQp2WU9XVmJjekF0MHFwTXZYRjNQWExqWmJVTTBrbk9kVUtJRWJxUDRZVWJnZHp4NlJ0Z2lpWTkzMEFqNnRBdGNlDQowZnBnTmx2ak1ScFNCdVdUbEFmTk5qRy9ZaG5kTXo5dUk2OFRNZkZwUjNQY2dWSXYzMGtydy85VnpvTGkyRHBlDQpvdzZEckdPNm9pK0RoTjc4UDRqWS9POVVjelpLMnJvWkwxT2k1UDBSSXhmMjNVWkM3eDFEbGNOM25CcjRzWVN2DQpyQng0Y0ZUTU5wd1UrbnpzSWk0ZGpjRkRLbUpkRU95ak1ua1AydjBMd2U3eXZLMDhwWmRFdSswemJycTE3a3VlDQpYcFhMYzdLNjhRQjE1eXh6R3lsVTVyUnd6bUMvWXNBVnlFNGVvR3U4UHhXeHJFUnZIYnk0QjhZUDB2QWZPcmFMDQpsS21YbEs0ZFRnPT0NCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0NCg==`

const validIntermediateAndCertRSA = `LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tDQpNSUlHTXpDQ0JCdWdBd0lCQWdJQkFUQU5CZ2txaGtpRzl3MEJBUXNGQURCZk1Rc3dDUVlEVlFRR0V3SlZVekVMDQpNQWtHQTFVRUNBd0NRMEV4RmpBVUJnTlZCQWNNRFZOaGJpQkdjbUZ1WTJselkyOHhEekFOQmdOVkJBb01Ca1J2DQpZMnRsY2pFYU1CZ0dBMVVFQXd3UlRtOTBZWEo1SUZSbGMzUnBibWNnUTBFd0hoY05NVFV3TnpFMk1EUXlOVEF6DQpXaGNOTWpVd056RXpNRFF5TlRBeldqQmZNUm93R0FZRFZRUUREQkZPYjNSaGNua2dWR1Z6ZEdsdVp5QkRRVEVMDQpNQWtHQTFVRUJoTUNWVk14RmpBVUJnTlZCQWNNRFZOaGJpQkdjbUZ1WTJselkyOHhEekFOQmdOVkJBb01Ca1J2DQpZMnRsY2pFTE1Ba0dBMVVFQ0F3Q1EwRXdnZ0lpTUEwR0NTcUdTSWIzRFFFQkFRVUFBNElDRHdBd2dnSUtBb0lDDQpBUUN3VlZENHBLN3o3cFhQcEpiYVoxSGc1ZVJYSWNhWXRiRlBDbk4waXF5OUhzVkVHbkVuNUJQTlNFc3VQK20wDQo1TjBxVlY3REdiMVNqaWxvTFhEMXFERHZoWFdrK2dpUzlwcHFQSFBMVlBCNGJ2enNxd0RZcnRwYnFrWXZPMFlLDQowU0wza3hQWFVGZGxrRmZndTB4amxjem0yUGhXRzNKZDhhQXRzcEwvTCtWZlBBMTNKVWFXeFNMcHVpMUluOHJoDQpnQXlRVEs2UTRPZjZHYkpZVG5BSGI1OVVvTFhTekI1QWZxaVVxNkw3bkVZWUtvUGZsUGJSQUlXTC9VQm0wYytIDQpvY21zNzA2UFlwbVBTMlJRdjNpT0dtbm45aEVWcDNQNmpxN1dBZXZiQTRhWUd4NUVzYlZ0WUFCcUpCYkZXQXV3DQp3VEdSWW16bjBNajBlVE1nZTl6dFlCMi8yc3hkVGU2dWhtRmdwVVhuZ0RxSkk1TzlOM3pQZnZsRUltQ2t5M0hNDQpqSm9MN2c1c21xWDlvMVArRVNMaDBWWnpoaDdJRFB6UVRYcGNQSVMvNnowbDIyUUdrSy8xTjFQYUFEYVVIZExMDQp2U2F2M3kyQmFFbVB2ZjJma1pqOHlQNWVZZ2k3Q3c1T05oSExEWUhGY2w5Wm0veXdtZHhISkVUejluZmdYbnNXDQpITnhEcXJrQ1ZPNDZyL3U2clNyVXQ2aHIzb2RkSkc4czhKbzA2ZWFydzZYVTNNek0rM2dpd2tLMFNTTTN1UlBxDQo0QXNjUjFUditFMzFBdU9BbWpxWVFvVDI5Yk1JeG9TemVsamovWW5lZHdqVzQ1cFd5YzNKb0hhaWJEd3ZXOVVvDQpHU1pCVnk0aHJNL0ZhN1hDV3YxV2ZITlcxZ0R3YUxZd0RubDVqRm1SQnZjZnVRSURBUUFCbzRINU1JSDJNSUdSDQpCZ05WSFNNRWdZa3dnWWFBRkhVTTFVM0U0V3lMMW52RmQrZFBZOGY0TzJoWm9XT2tZVEJmTVFzd0NRWURWUVFHDQpFd0pWVXpFTE1Ba0dBMVVFQ0F3Q1EwRXhGakFVQmdOVkJBY01EVk5oYmlCR2NtRnVZMmx6WTI4eER6QU5CZ05WDQpCQW9NQmtSdlkydGxjakVhTUJnR0ExVUVBd3dSVG05MFlYSjVJRlJsYzNScGJtY2dRMEdDQ1FEQ2VETGJlbUlUDQpTekFTQmdOVkhSTUJBZjhFQ0RBR0FRSC9BZ0VBTUIwR0ExVWRKUVFXTUJRR0NDc0dBUVVGQndNQ0JnZ3JCZ0VGDQpCUWNEQVRBT0JnTlZIUThCQWY4RUJBTUNBVVl3SFFZRFZSME9CQllFRkhlNDhoY0JjQXAwYlVWbFR4WGVSQTRvDQpFMTZwTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElDQVFBV1V0QVBkVUZwd1JxK04xU3pHVWVqU2lrZU1HeVBac2NaDQpKQlVDbWhab0Z1ZmdYR2JMTzVPcGNSTGFWM1hkYTB0LzVQdGRHTVNFemN6ZW9aSFdrbkR0dys3OU9CaXR0UFBqDQpTaDFvRkR1UG8zNVI3ZVA2MjRsVUNjaC9JblpDcGhUYUx4OW9ETEdjYUszYWlsUTl3akJkS2RsQmw4S05LSVpwDQphMTNhUDVyblNtMkp2YSt0WHkveWkzQlNkczNkR0Q4SVRLWnlJLzZBRkh4R3ZPYnJESUJwbzRGRi96Y1dYVkRqDQpwYU9teHBsUnRNNEhpdG0rc1hHdmZxSmU0eDVEdU9YT25QclQzZEh2UlQ2dlNaVW9Lb2J4TXFtUlRPY3JPSVBhDQpFZU1wT29ic2hPUnVSbnRNRFl2dmdPM0Q2cDZpY2lEVzJWcDlONnJkTWRmT1dFUU44SlZXdkI3SXhSSGs5cUtKDQp2WU9XVmJjekF0MHFwTXZYRjNQWExqWmJVTTBrbk9kVUtJRWJxUDRZVWJnZHp4NlJ0Z2lpWTkzMEFqNnRBdGNlDQowZnBnTmx2ak1ScFNCdVdUbEFmTk5qRy9ZaG5kTXo5dUk2OFRNZkZwUjNQY2dWSXYzMGtydy85VnpvTGkyRHBlDQpvdzZEckdPNm9pK0RoTjc4UDRqWS9POVVjelpLMnJvWkwxT2k1UDBSSXhmMjNVWkM3eDFEbGNOM25CcjRzWVN2DQpyQng0Y0ZUTU5wd1UrbnpzSWk0ZGpjRkRLbUpkRU95ak1ua1AydjBMd2U3eXZLMDhwWmRFdSswemJycTE3a3VlDQpYcFhMYzdLNjhRQjE1eXh6R3lsVTVyUnd6bUMvWXNBVnlFNGVvR3U4UHhXeHJFUnZIYnk0QjhZUDB2QWZPcmFMDQpsS21YbEs0ZFRnPT0NCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0NCi0tLS0tQkVHSU4gQ0VSVElGSUNBVEUtLS0tLQ0KTUlJRlZ6Q0NBeitnQXdJQkFnSUJBekFOQmdrcWhraUc5dzBCQVFzRkFEQmZNUm93R0FZRFZRUUREQkZPYjNSaA0KY25rZ1ZHVnpkR2x1WnlCRFFURUxNQWtHQTFVRUJoTUNWVk14RmpBVUJnTlZCQWNNRFZOaGJpQkdjbUZ1WTJseg0KWTI4eER6QU5CZ05WQkFvTUJrUnZZMnRsY2pFTE1Ba0dBMVVFQ0F3Q1EwRXdIaGNOTVRVd056RTJNRFF5TlRVdw0KV2hjTk1UWXdOekUxTURReU5UVXdXakJnTVJzd0dRWURWUVFEREJKelpXTjFjbVV1WlhoaGJYQnNaUzVqYjIweA0KQ3pBSkJnTlZCQVlUQWxWVE1SWXdGQVlEVlFRSERBMVRZVzRnUm5KaGJtTnBjMk52TVE4d0RRWURWUVFLREFaRQ0KYjJOclpYSXhDekFKQmdOVkJBZ01Ba05CTUlJQklqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FROEFNSUlCQ2dLQw0KQVFFQW1MWWlZQ1RBV0pCV0F1eFpMcVZtVjRGaVVkR2dFcW9RdkNiTjczekYvbVFmaHEwQ0lUbzZ4U3hzMVFpRw0KRE96VXRrcHpYenppU2o0SjUrZXQ0SmtGbGVlRUthTWNIYWRlSXNTbEhHdlZ0WER2OTNvUjN5ZG1mWk8rVUxSVQ0KOHhIbG9xY0xyMUtyT1AxZGFMZmRNUmJhY3RkNzVVUWd2dzlYVHNkZU1WWDVBbGljU0VOVktWK0FRWHZWcHY4UA0KVDEwTVN2bEJGYW00cmVYdVkvU2tlTWJJYVc1cEZ1NkFRdjNabWZ0dDJ0YTBDQjlrYjFtWWQrT0tydThIbm5xNQ0KYUp3NlIzR2hQMFRCZDI1UDFQa2lTeE0yS0dZWlprMFcvTlpxTEs5L0xURktUTkN2N1ZqQ2J5c1ZvN0h4Q1kwYg0KUWUvYkRQODJ2N1NuTHRiM2Fab2dmdmE0SFFJREFRQUJvNElCR3pDQ0FSY3dnWWdHQTFVZEl3U0JnREIrZ0JSMw0KdVBJWEFYQUtkRzFGWlU4VjNrUU9LQk5lcWFGanBHRXdYekVMTUFrR0ExVUVCaE1DVlZNeEN6QUpCZ05WQkFnTQ0KQWtOQk1SWXdGQVlEVlFRSERBMVRZVzRnUm5KaGJtTnBjMk52TVE4d0RRWURWUVFLREFaRWIyTnJaWEl4R2pBWQ0KQmdOVkJBTU1FVTV2ZEdGeWVTQlVaWE4wYVc1bklFTkJnZ0VCTUF3R0ExVWRFd0VCL3dRQ01BQXdIUVlEVlIwbA0KQkJZd0ZBWUlLd1lCQlFVSEF3SUdDQ3NHQVFVRkJ3TUJNQTRHQTFVZER3RUIvd1FFQXdJRm9EQXVCZ05WSFJFRQ0KSnpBbGdoSnpaV04xY21VdVpYaGhiWEJzWlM1amIyMkNDV3h2WTJGc2FHOXpkSWNFZndBQUFUQWRCZ05WSFE0RQ0KRmdRVURQRDRDYVhSYnU1UUJiNWU4eThvZHZUcVc0SXdEUVlKS29aSWh2Y05BUUVMQlFBRGdnSUJBSk95bG1jNA0KbjdKNjRHS3NQL3hoVWRLS1Y5L0tEK3VmenBLYnJMSW9qV243clR5ZTcwdlkwT2pRRnVPWGM1NHlqTVNJTCsvNQ0KbWxOUTdZL2ZKUzh4ZEg3OUVSKzRuV011RDJlY2lMbnNMZ2JZVWs0aGl5Ynk4LzVWKy9ZcVBlQ3BQQ242VEpSSw0KYTBFNmxWL1VqWEpkcmlnSnZKb05PUjhaZ3RFWi9RUGdqSkVWVXNnNDdkdHF6c0RwZ2VTOGRjanVNV3BaeFAwMg0KcWF2RkxEalNGelZIKzJENk90eTFEUXBsbS8vM1hhUlhoMjNkT0NQOHdqL2J4dm5WVG9GV3Mrek80dVQxTEYvUw0KS1hDTlFvZWlHeFdIeXpyWEZWVnRWbkM5RlNOejBHZzIvRW0xdGZSZ3ZoVW40S0xKY3ZaVzlvMVI3VlZDWDBMMQ0KMHgwZnlLM1ZXZVdjODZhNWE2ODFhbUtaU0ViakFtSVZaRjl6T1gwUE9EQzhveSt6cU9QV2EwV0NsNEs2ekRDNg0KMklJRkJCTnk1MFpTMmlPTjZSWTZtRTdObUE3OGdja2Y0MTVjcUlWcmxvWUpiYlREZXBmaFRWMjE4U0xlcHBoNA0KdUdiMi9zeGtsZkhPWUUrcnBIY2lpYld3WHJ3bE9ESmFYdXpYRmhwbFVkL292ZHVqQk5BSUhrQmZ6eStZNnoycw0KYndaY2ZxRDROSWIvQUdoSXlXMnZidnU0enNsRHAxTUVzTG9hTytTemlyTXpreU1CbEtSdDEyMHR3czRFa1VsbQ0KL1FoalNVb1pwQ0FzeTVDL3BWNCtieDBTeXNOZC9TK2tLYVJaYy9VNlkzWllCRmhzekxoN0phTFhLbWs3d0huRQ0KcmdnbTZvejRML0d5UFdjL0ZqZm5zZWZXS00yeUMzUURoanZqDQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tDQo=`

const signedRSARootTemplate = `{"signed":{"_type":"Root","consistent_snapshot":false,"expires":"2016-07-16T23:34:13.389129622-07:00","keys":{"1fc4fdc38f66558658c5c59b67f1716bdc6a74ef138b023ae5931db69f51d670":{"keytype":"ecdsa","keyval":{"private":null,"public":"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE8nIgzLigo5D47dWQe1IUjzHXxvyx0j/OL16VQymuloWsgVDxxT6+mH3CeviMAs+/McnEPE9exnm6SQGR5x3XMw=="}},"23c29cc372109c819e081bc953b7657d05e3f968f03c21d0d75ea457590f3d14":{"keytype":"ecdsa","keyval":{"private":null,"public":"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEClUFVWkc85OQScfTQRS02VaLIEaeCmxdwYS/hcTLVoTxlFfRfs7HyalTwXGAGO79XZZS+koE6s8D0xGcCJQkLQ=="}},"49cf5c6404a35fa41d5a5aa2ce539dfee0d7a2176d0da488914a38603b1f4292":{"keytype":"rsa-x509","keyval":{"private":null,"public":"{{.RootPem}}"}},"e3a5a4fdaf11ea1ec58f5efed6f3639b39cd4cfa1418c8b55c9a8c2447ace5d9":{"keytype":"ecdsa","keyval":{"private":null,"public":"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEgl3rzMPMEKhS1k/AX16MM4PdidpjJr+z4pj0Td+30QnpbOIARgpyR1PiFztU8BZlqG3cUazvFclr2q/xHvfrqw=="}}},"roles":{"root":{"keyids":["49cf5c6404a35fa41d5a5aa2ce539dfee0d7a2176d0da488914a38603b1f4292"],"threshold":1},"snapshot":{"keyids":["23c29cc372109c819e081bc953b7657d05e3f968f03c21d0d75ea457590f3d14"],"threshold":1},"targets":{"keyids":["1fc4fdc38f66558658c5c59b67f1716bdc6a74ef138b023ae5931db69f51d670"],"threshold":1},"timestamp":{"keyids":["e3a5a4fdaf11ea1ec58f5efed6f3639b39cd4cfa1418c8b55c9a8c2447ace5d9"],"threshold":1}},"version":2},"signatures":[{"keyid":"49cf5c6404a35fa41d5a5aa2ce539dfee0d7a2176d0da488914a38603b1f4292","method":"rsapss","sig":"YlZwtCj028Xc23+KHfj6govFEY6hMbBXO5HT20F0I5ZeIPb1l7OmkjEiwp9ZHusClY+QeqiP1CFh\n/AfCbv4tLanqMkXPtm8UJJ1hMZVq86coieB32PQDj9k6x1hErHzvPUbOzTRW2BQkFFMZFkLDAd06\npH8lmxyPLOhdkVE8qIT7sBCy/4bQIGfvEX6yCDz84MZdcLNX5B9mzGi9A7gDloh9IEZxA8UgoI18\nSYpv/fYeSZSqM/ws2G+kiELGgTWhcZ+gOlF7ArM/DOlcC/NYqcvY1ugE6Gn7G8opre6NOofdRp3w\n603A2rMMvYTwqKLY6oX/d+07A2+WGHXPUy5otCAybWOw2hIZ35Jjmh12g6Dc6Qk4K2zXwAgvWwBU\nWlT8MlP1Tf7f80jnGjh0aARlHI4LCxlYU5L/pCaYuHgynujvLuzoOuiiPfJv7sYvKoQ8UieE1w//\nHc8E6tWtV5G2FguKLurMoKZ9FBWcanDO0fg5AWuG3qcgUJdvh9acQ33EKer1fqBxs6LSAUWo8rDt\nQkg+b55AW0YBukAW9IAfMySQGAS2e3mHZ8nK/ijaygCRu7/P+NgKY9/zpmfL2xgcNslLcANcSOOt\nhiJS6yqYM9i9G0af0yw/TxAT4ntwjVm8u52UyR/hXIiUc/mjZcYRbSmJOHws902+i+Z/qv72knk="}]}`

const rootPubKeyID = "49cf5c6404a35fa41d5a5aa2ce539dfee0d7a2176d0da488914a38603b1f4292"

const targetsPubKeyID = "1fc4fdc38f66558658c5c59b67f1716bdc6a74ef138b023ae5931db69f51d670"

func TestValidateRoot(t *testing.T) {
	var testSignedRoot data.Signed
	var signedRootBytes bytes.Buffer

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	defer os.RemoveAll(tempBaseDir)
	require.NoError(t, err, "failed to create a temporary directory: %s", err)

	// Execute our template
	templ, _ := template.New("SignedRSARootTemplate").Parse(signedRSARootTemplate)
	templ.Execute(&signedRootBytes, SignedRSARootTemplate{RootPem: validPEMEncodedRSARoot})

	// Unmarshal our signedroot
	json.Unmarshal(signedRootBytes.Bytes(), &testSignedRoot)

	// This call to trustpinning.ValidateRoot will succeed since we are using a valid PEM
	// encoded certificate, and have no other certificates for this CN
	_, err = trustpinning.ValidateRoot(nil, &testSignedRoot, "docker.com/notary", trustpinning.TrustPinConfig{})
	require.NoError(t, err)

	// This call to trustpinning.ValidateRoot will fail since we are passing in a dnsName that
	// doesn't match the CN of the certificate.
	_, err = trustpinning.ValidateRoot(nil, &testSignedRoot, "diogomonica.com/notary", trustpinning.TrustPinConfig{})
	require.Error(t, err, "An error was expected")
	require.Equal(t, err, &trustpinning.ErrValidationFail{Reason: "unable to retrieve valid leaf certificates"})

	//
	// This call to trustpinning.ValidateRoot will fail since we are passing an unparsable RootSigned
	//
	// Execute our template deleting the old buffer first
	signedRootBytes.Reset()
	templ, _ = template.New("SignedRSARootTemplate").Parse(signedRSARootTemplate)
	templ.Execute(&signedRootBytes, SignedRSARootTemplate{RootPem: "------ ABSOLUTELY NOT A PEM -------"})
	// Unmarshal our signedroot
	json.Unmarshal(signedRootBytes.Bytes(), &testSignedRoot)

	_, err = trustpinning.ValidateRoot(nil, &testSignedRoot, "docker.com/notary", trustpinning.TrustPinConfig{})
	require.Error(t, err, "illegal base64 data at input byte")

	//
	// This call to trustpinning.ValidateRoot will fail since we are passing an invalid PEM cert
	//
	// Execute our template deleting the old buffer first
	signedRootBytes.Reset()
	templ, _ = template.New("SignedRSARootTemplate").Parse(signedRSARootTemplate)
	templ.Execute(&signedRootBytes, SignedRSARootTemplate{RootPem: "LS0tLS1CRUdJTiBDRVJU"})
	// Unmarshal our signedroot
	json.Unmarshal(signedRootBytes.Bytes(), &testSignedRoot)

	_, err = trustpinning.ValidateRoot(nil, &testSignedRoot, "docker.com/notary", trustpinning.TrustPinConfig{})
	require.Error(t, err, "An error was expected")
	require.Equal(t, err, &trustpinning.ErrValidationFail{Reason: "unable to retrieve valid leaf certificates"})

	//
	// This call to trustpinning.ValidateRoot will fail since we are passing only CA certificate
	// This will fail due to the lack of a leaf certificate
	//
	// Execute our template deleting the old buffer first
	signedRootBytes.Reset()
	templ, _ = template.New("SignedRSARootTemplate").Parse(signedRSARootTemplate)
	templ.Execute(&signedRootBytes, SignedRSARootTemplate{RootPem: validCAPEMEncodeRSARoot})
	// Unmarshal our signedroot
	json.Unmarshal(signedRootBytes.Bytes(), &testSignedRoot)

	_, err = trustpinning.ValidateRoot(nil, &testSignedRoot, "docker.com/notary", trustpinning.TrustPinConfig{})
	require.Error(t, err, "An error was expected")
	require.Equal(t, err, &trustpinning.ErrValidationFail{Reason: "unable to retrieve valid leaf certificates"})

	//
	// This call to trustpinning.ValidateRoot could succeed in getting to the TUF validation, since
	// we are using a valid PEM encoded certificate chain of intermediate + leaf cert
	// that are signed by a trusted root authority and the leaf cert has a correct CN.
	// It will, however, fail to validate, because the leaf cert does not precede the
	// intermediate in the certificate bundle
	//
	// Execute our template deleting the old buffer first
	signedRootBytes.Reset()
	templ, _ = template.New("SignedRSARootTemplate").Parse(signedRSARootTemplate)
	templ.Execute(&signedRootBytes, SignedRSARootTemplate{RootPem: validIntermediateAndCertRSA})

	// Unmarshal our signedroot
	json.Unmarshal(signedRootBytes.Bytes(), &testSignedRoot)

	_, err = trustpinning.ValidateRoot(nil, &testSignedRoot, "secure.example.com", trustpinning.TrustPinConfig{})
	require.Error(t, err, "An error was expected")
	require.Equal(t, err, &trustpinning.ErrValidationFail{Reason: "unable to retrieve valid leaf certificates"})
}

func TestValidateRootWithoutTOFUS(t *testing.T) {
	var testSignedRoot data.Signed
	var signedRootBytes bytes.Buffer

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	defer os.RemoveAll(tempBaseDir)
	require.NoError(t, err, "failed to create a temporary directory: %s", err)

	// Execute our template
	templ, _ := template.New("SignedRSARootTemplate").Parse(signedRSARootTemplate)
	templ.Execute(&signedRootBytes, SignedRSARootTemplate{RootPem: validPEMEncodedRSARoot})

	// Unmarshal our signedroot
	json.Unmarshal(signedRootBytes.Bytes(), &testSignedRoot)

	// This call to trustpinning.ValidateRoot will fail since we are explicitly disabling TOFU and have no local certs
	_, err = trustpinning.ValidateRoot(nil, &testSignedRoot, "docker.com/notary", trustpinning.TrustPinConfig{DisableTOFU: true})
	require.Error(t, err)
}

func TestValidateRootWithPinnedCert(t *testing.T) {
	var testSignedRoot data.Signed
	var signedRootBytes bytes.Buffer

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	defer os.RemoveAll(tempBaseDir)
	require.NoError(t, err, "failed to create a temporary directory: %s", err)

	// Execute our template
	templ, _ := template.New("SignedRSARootTemplate").Parse(signedRSARootTemplate)
	templ.Execute(&signedRootBytes, SignedRSARootTemplate{RootPem: validPEMEncodedRSARoot})

	// Unmarshal our signedroot
	json.Unmarshal(signedRootBytes.Bytes(), &testSignedRoot)
	typedSignedRoot, err := data.RootFromSigned(&testSignedRoot)
	require.NoError(t, err)

	// This call to trustpinning.ValidateRoot should succeed with the correct Cert ID (same as root public key ID)
	validatedSignedRoot, err := trustpinning.ValidateRoot(nil, &testSignedRoot, "docker.com/notary", trustpinning.TrustPinConfig{Certs: map[string][]string{"docker.com/notary": {rootPubKeyID}}, DisableTOFU: true})
	require.NoError(t, err)
	typedSignedRoot.Signatures[0].IsValid = true
	require.Equal(t, validatedSignedRoot, typedSignedRoot)

	// This call to trustpinning.ValidateRoot should also succeed with the correct Cert ID (same as root public key ID), even though we passed an extra bad one
	validatedSignedRoot, err = trustpinning.ValidateRoot(nil, &testSignedRoot, "docker.com/notary", trustpinning.TrustPinConfig{Certs: map[string][]string{"docker.com/notary": {rootPubKeyID, "invalidID"}}, DisableTOFU: true})
	require.NoError(t, err)
	// This extra assignment is necessary because ValidateRoot calls through to a successful VerifySignature which marks IsValid
	typedSignedRoot.Signatures[0].IsValid = true
	require.Equal(t, typedSignedRoot, validatedSignedRoot)
}

func TestValidateRootWithPinnedCertAndIntermediates(t *testing.T) {
	now := time.Now()
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	memStore := trustmanager.NewKeyMemoryStore(passphraseRetriever)
	cs := cryptoservice.NewCryptoService(memStore)

	// generate CA cert
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	require.NoError(t, err)
	caTmpl := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "notary testing CA",
		},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:       true,
		MaxPathLen: 3,
	}
	caPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	_, err = x509.CreateCertificate(
		rand.Reader,
		&caTmpl,
		&caTmpl,
		caPrivKey.Public(),
		caPrivKey,
	)
	require.NoError(t, err)

	// generate intermediate
	intTmpl := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "notary testing intermediate",
		},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:       true,
		MaxPathLen: 2,
	}
	intPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	intCert, err := x509.CreateCertificate(
		rand.Reader,
		&intTmpl,
		&caTmpl,
		intPrivKey.Public(),
		caPrivKey,
	)
	require.NoError(t, err)

	// generate leaf
	serialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
	require.NoError(t, err)
	leafTmpl := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "docker.io/notary/test",
		},
		NotBefore: now.Add(-time.Hour),
		NotAfter:  now.Add(time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
		BasicConstraintsValid: true,
	}

	leafPubKey, err := cs.Create("root", "docker.io/notary/test", data.ECDSAKey)
	require.NoError(t, err)
	leafPrivKey, _, err := cs.GetPrivateKey(leafPubKey.ID())
	require.NoError(t, err)
	signer := leafPrivKey.CryptoSigner()
	leafCert, err := x509.CreateCertificate(
		rand.Reader,
		&leafTmpl,
		&intTmpl,
		signer.Public(),
		intPrivKey,
	)
	require.NoError(t, err)

	rootBundleWriter := bytes.NewBuffer(nil)
	pem.Encode(
		rootBundleWriter,
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: leafCert,
		},
	)
	pem.Encode(
		rootBundleWriter,
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: intCert,
		},
	)

	rootBundle := rootBundleWriter.Bytes()

	ecdsax509Key := data.NewECDSAx509PublicKey(rootBundle)

	otherKey, err := cs.Create("targets", "docker.io/notary/test", data.ED25519Key)
	require.NoError(t, err)

	root := data.SignedRoot{
		Signatures: make([]data.Signature, 0),
		Signed: data.Root{
			SignedCommon: data.SignedCommon{
				Type:    "Root",
				Expires: now.Add(time.Hour),
				Version: 1,
			},
			Keys: map[string]data.PublicKey{
				ecdsax509Key.ID(): ecdsax509Key,
				otherKey.ID():     otherKey,
			},
			Roles: map[data.RoleName]*data.RootRole{
				"root": {
					KeyIDs:    []string{ecdsax509Key.ID()},
					Threshold: 1,
				},
				"targets": {
					KeyIDs:    []string{otherKey.ID()},
					Threshold: 1,
				},
				"snapshot": {
					KeyIDs:    []string{otherKey.ID()},
					Threshold: 1,
				},
				"timestamp": {
					KeyIDs:    []string{otherKey.ID()},
					Threshold: 1,
				},
			},
		},
		Dirty: true,
	}

	signedRoot, err := root.ToSigned()
	require.NoError(t, err)
	err = signed.Sign(cs, signedRoot, []data.PublicKey{ecdsax509Key}, 1, nil)
	require.NoError(t, err)

	typedSignedRoot, err := data.RootFromSigned(signedRoot)
	require.NoError(t, err)

	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	defer os.RemoveAll(tempBaseDir)
	require.NoError(t, err, "failed to create a temporary directory: %s", err)

	validatedRoot, err := trustpinning.ValidateRoot(
		nil,
		signedRoot,
		"docker.io/notary/test",
		trustpinning.TrustPinConfig{
			Certs: map[string][]string{
				"docker.io/notary/test": {ecdsax509Key.ID()},
			},
			DisableTOFU: true,
		},
	)
	require.NoError(t, err, "failed to validate certID with intermediate")
	for idx, sig := range typedSignedRoot.Signatures {
		if sig.KeyID == ecdsax509Key.ID() {
			typedSignedRoot.Signatures[idx].IsValid = true
		}
	}
	require.Equal(t, typedSignedRoot, validatedRoot)
}

func TestValidateRootFailuresWithPinnedCert(t *testing.T) {
	var testSignedRoot data.Signed
	var signedRootBytes bytes.Buffer

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	defer os.RemoveAll(tempBaseDir)
	require.NoError(t, err, "failed to create a temporary directory: %s", err)

	// Execute our template
	templ, _ := template.New("SignedRSARootTemplate").Parse(signedRSARootTemplate)
	templ.Execute(&signedRootBytes, SignedRSARootTemplate{RootPem: validPEMEncodedRSARoot})

	// Unmarshal our signedroot
	json.Unmarshal(signedRootBytes.Bytes(), &testSignedRoot)
	typedSignedRoot, err := data.RootFromSigned(&testSignedRoot)
	require.NoError(t, err)

	// This call to trustpinning.ValidateRoot should fail due to an incorrect cert ID
	_, err = trustpinning.ValidateRoot(nil, &testSignedRoot, "docker.com/notary", trustpinning.TrustPinConfig{Certs: map[string][]string{"docker.com/notary": {"ABSOLUTELY NOT A CERT ID"}}, DisableTOFU: true})
	require.Error(t, err)

	// This call to trustpinning.ValidateRoot should fail due to an empty cert ID
	_, err = trustpinning.ValidateRoot(nil, &testSignedRoot, "docker.com/notary", trustpinning.TrustPinConfig{Certs: map[string][]string{"docker.com/notary": {""}}, DisableTOFU: true})
	require.Error(t, err)

	// This call to trustpinning.ValidateRoot should fail due to an invalid GUN (even though the cert ID is correct), and TOFUS is set to false
	_, err = trustpinning.ValidateRoot(nil, &testSignedRoot, "docker.com/notary", trustpinning.TrustPinConfig{Certs: map[string][]string{"not_a_gun": {rootPubKeyID}}, DisableTOFU: true})
	require.Error(t, err)

	// This call to trustpinning.ValidateRoot should fail due to an invalid cert ID, even though it's a valid key ID for targets
	_, err = trustpinning.ValidateRoot(nil, &testSignedRoot, "docker.com/notary", trustpinning.TrustPinConfig{Certs: map[string][]string{"docker.com/notary": {targetsPubKeyID}}, DisableTOFU: true})
	require.Error(t, err)

	// This call to trustpinning.ValidateRoot should succeed because we fall through to TOFUS because we have no matching GUNs under Certs
	validatedRoot, err := trustpinning.ValidateRoot(nil, &testSignedRoot, "docker.com/notary", trustpinning.TrustPinConfig{Certs: map[string][]string{"not_a_gun": {rootPubKeyID}}, DisableTOFU: false})
	require.NoError(t, err)
	typedSignedRoot.Signatures[0].IsValid = true
	require.Equal(t, typedSignedRoot, validatedRoot)
}

func TestValidateRootWithPinnedCA(t *testing.T) {
	var testSignedRoot data.Signed
	var signedRootBytes bytes.Buffer

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	defer os.RemoveAll(tempBaseDir)
	require.NoError(t, err, "failed to create a temporary directory: %s", err)

	templ, _ := template.New("SignedRSARootTemplate").Parse(signedRSARootTemplate)
	templ.Execute(&signedRootBytes, SignedRSARootTemplate{RootPem: validPEMEncodedRSARoot})
	// Unmarshal our signedRoot
	json.Unmarshal(signedRootBytes.Bytes(), &testSignedRoot)
	typedSignedRoot, err := data.RootFromSigned(&testSignedRoot)
	require.NoError(t, err)

	// This call to trustpinning.ValidateRoot will fail because we have an invalid path for the CA
	_, err = trustpinning.ValidateRoot(nil, &testSignedRoot, "docker.com/notary", trustpinning.TrustPinConfig{CA: map[string]string{"docker.com/notary": filepath.Join(tempBaseDir, "nonexistent")}})
	require.Error(t, err)

	// This call to trustpinning.ValidateRoot will fail because we have no valid GUNs to use, and TOFUS is disabled
	_, err = trustpinning.ValidateRoot(nil, &testSignedRoot, "docker.com/notary", trustpinning.TrustPinConfig{CA: map[string]string{"othergun": filepath.Join(tempBaseDir, "nonexistent")}, DisableTOFU: true})
	require.Error(t, err)

	// This call to trustpinning.ValidateRoot will succeed because we have no valid GUNs to use and we fall back to enabled TOFUS
	validatedRoot, err := trustpinning.ValidateRoot(nil, &testSignedRoot, "docker.com/notary", trustpinning.TrustPinConfig{CA: map[string]string{"othergun": filepath.Join(tempBaseDir, "nonexistent")}, DisableTOFU: false})
	require.NoError(t, err)
	typedSignedRoot.Signatures[0].IsValid = true
	require.Equal(t, typedSignedRoot, validatedRoot)

	// Write an invalid CA cert (not even a PEM) to the tempDir and ensure validation fails when using it
	invalidCAFilepath := filepath.Join(tempBaseDir, "invalid.ca")
	require.NoError(t, ioutil.WriteFile(invalidCAFilepath, []byte("ABSOLUTELY NOT A PEM"), 0644))

	// Using this invalid CA cert should fail on trustpinning.ValidateRoot
	_, err = trustpinning.ValidateRoot(nil, &testSignedRoot, "docker.com/notary", trustpinning.TrustPinConfig{CA: map[string]string{"docker.com/notary": invalidCAFilepath}, DisableTOFU: true})
	require.Error(t, err)

	validCAFilepath := "../fixtures/root-ca.crt"

	// If we pass an invalid Certs entry in addition to this valid CA entry, since Certs has priority for pinning we will fail
	_, err = trustpinning.ValidateRoot(nil, &testSignedRoot, "docker.com/notary", trustpinning.TrustPinConfig{Certs: map[string][]string{"docker.com/notary": {"invalidID"}}, CA: map[string]string{"docker.com/notary": validCAFilepath}, DisableTOFU: true})
	require.Error(t, err)

	// Now construct a new root with a valid cert chain, such that signatures are correct over the 'notary-signer' GUN.  Pin the root-ca and validate
	certChain, err := utils.LoadCertBundleFromFile("../fixtures/notary-signer.crt")
	require.NoError(t, err)

	pemChainBytes, err := utils.CertChainToPEM(certChain)
	require.NoError(t, err)

	newRootKey := data.NewPublicKey(data.RSAx509Key, pemChainBytes)

	rootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{newRootKey.ID()}, nil)
	require.NoError(t, err)

	testRoot, err := data.NewRoot(
		map[string]data.PublicKey{newRootKey.ID(): newRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &rootRole.RootRole,
			data.CanonicalTimestampRole: &rootRole.RootRole,
			data.CanonicalTargetsRole:   &rootRole.RootRole,
			data.CanonicalSnapshotRole:  &rootRole.RootRole},
		false,
	)
	testRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	keyReader, err := os.Open("../fixtures/notary-signer.key")
	require.NoError(t, err, "could not open key file")
	pemBytes, err := ioutil.ReadAll(keyReader)
	require.NoError(t, err, "could not read key file")
	privKey, err := utils.ParsePEMPrivateKey(pemBytes, "")
	require.NoError(t, err)

	store, err := trustmanager.NewKeyFileStore(tempBaseDir, passphraseRetriever)
	require.NoError(t, err)
	cs := cryptoservice.NewCryptoService(store)

	err = store.AddKey(trustmanager.KeyInfo{Role: data.CanonicalRootRole, Gun: "notary-signer"}, privKey)
	require.NoError(t, err)

	newTestSignedRoot, err := testRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, newTestSignedRoot, []data.PublicKey{newRootKey}, 1, nil)
	require.NoError(t, err)

	newTypedSignedRoot, err := data.RootFromSigned(newTestSignedRoot)
	require.NoError(t, err)

	// Check that we validate correctly against a pinned CA and provided bundle
	validatedRoot, err = trustpinning.ValidateRoot(nil, newTestSignedRoot, "notary-signer", trustpinning.TrustPinConfig{CA: map[string]string{"notary-signer": validCAFilepath}, DisableTOFU: true})
	require.NoError(t, err)
	for idx, sig := range newTypedSignedRoot.Signatures {
		if sig.KeyID == newRootKey.ID() {
			newTypedSignedRoot.Signatures[idx].IsValid = true
		}
	}
	require.Equal(t, newTypedSignedRoot, validatedRoot)

	// Add an expired CA for the same gun to our previous pinned bundle, ensure that we still validate correctly
	goodRootCABundle, err := utils.LoadCertBundleFromFile(validCAFilepath)
	require.NoError(t, err)
	memKeyStore := trustmanager.NewKeyMemoryStore(passphraseRetriever)
	cryptoService := cryptoservice.NewCryptoService(memKeyStore)
	testPubKey, err := cryptoService.Create("root", "notary-signer", data.ECDSAKey)
	require.NoError(t, err)
	testPrivKey, _, err := memKeyStore.GetKey(testPubKey.ID())
	require.NoError(t, err)
	expiredCert, err := generateExpiredTestingCertificate(testPrivKey, "notary-signer")
	require.NoError(t, err)
	bundleWithExpiredCert, err := utils.CertChainToPEM(append(goodRootCABundle, expiredCert))
	require.NoError(t, err)
	bundleWithExpiredCertPath := filepath.Join(tempBaseDir, "bundle_with_expired_cert.pem")
	require.NoError(t, ioutil.WriteFile(bundleWithExpiredCertPath, bundleWithExpiredCert, 0644))

	// Check that we validate correctly against a pinned CA and provided bundle
	validatedRoot, err = trustpinning.ValidateRoot(nil, newTestSignedRoot, "notary-signer", trustpinning.TrustPinConfig{CA: map[string]string{"notary-signer": bundleWithExpiredCertPath}, DisableTOFU: true})
	require.NoError(t, err)
	require.Equal(t, newTypedSignedRoot, validatedRoot)

	testPubKey2, err := cryptoService.Create("root", "notary-signer", data.ECDSAKey)
	require.NoError(t, err)
	testPrivKey2, _, err := memKeyStore.GetKey(testPubKey2.ID())
	require.NoError(t, err)
	expiredCert2, err := generateExpiredTestingCertificate(testPrivKey2, "notary-signer")
	require.NoError(t, err)
	allExpiredCertBundle, err := utils.CertChainToPEM([]*x509.Certificate{expiredCert, expiredCert2})
	require.NoError(t, err)
	allExpiredCertPath := filepath.Join(tempBaseDir, "all_expired_cert.pem")
	require.NoError(t, ioutil.WriteFile(allExpiredCertPath, allExpiredCertBundle, 0644))
	// Now only use expired certs in the bundle, we should fail
	_, err = trustpinning.ValidateRoot(nil, newTestSignedRoot, "notary-signer", trustpinning.TrustPinConfig{CA: map[string]string{"notary-signer": allExpiredCertPath}, DisableTOFU: true})
	require.Error(t, err)

	// Add a CA cert for a that won't validate against the root leaf certificate
	testPubKey3, err := cryptoService.Create("root", "notary-signer", data.ECDSAKey)
	require.NoError(t, err)
	testPrivKey3, _, err := memKeyStore.GetKey(testPubKey3.ID())
	require.NoError(t, err)
	validCert, err := cryptoservice.GenerateCertificate(testPrivKey3, "notary-signer", time.Now(), time.Now().AddDate(1, 0, 0))
	require.NoError(t, err)
	bundleWithWrongCert, err := utils.CertChainToPEM([]*x509.Certificate{validCert})
	require.NoError(t, err)
	bundleWithWrongCertPath := filepath.Join(tempBaseDir, "bundle_with_expired_cert.pem")
	require.NoError(t, ioutil.WriteFile(bundleWithWrongCertPath, bundleWithWrongCert, 0644))
	_, err = trustpinning.ValidateRoot(nil, newTestSignedRoot, "notary-signer", trustpinning.TrustPinConfig{CA: map[string]string{"notary-signer": bundleWithWrongCertPath}, DisableTOFU: true})
	require.Error(t, err)
}

// TestValidateSuccessfulRootRotation runs through a full root certificate rotation
// We test this with both an RSA and ECDSA root certificate
func TestValidateSuccessfulRootRotation(t *testing.T) {
	testValidateSuccessfulRootRotation(t, data.ECDSAKey, data.ECDSAx509Key)
	if !testing.Short() {
		testValidateSuccessfulRootRotation(t, data.RSAKey, data.RSAx509Key)
	}
}

func testValidateSuccessfulRootRotation(t *testing.T, keyAlg, rootKeyType string) {
	// The gun to test
	var gun data.GUN = "docker.com/notary"

	memKeyStore := trustmanager.NewKeyMemoryStore(passphraseRetriever)
	cs := cryptoservice.NewCryptoService(memKeyStore)

	// TUF key with PEM-encoded x509 certificate
	origRootKey, err := testutils.CreateKey(cs, gun, data.CanonicalRootRole, keyAlg)
	require.NoError(t, err)

	origRootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{origRootKey.ID()}, nil)
	require.NoError(t, err)

	origTestRoot, err := data.NewRoot(
		map[string]data.PublicKey{origRootKey.ID(): origRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &origRootRole.RootRole,
			data.CanonicalTargetsRole:   &origRootRole.RootRole,
			data.CanonicalSnapshotRole:  &origRootRole.RootRole,
			data.CanonicalTimestampRole: &origRootRole.RootRole,
		},
		false,
	)
	origTestRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	signedOrigTestRoot, err := origTestRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, signedOrigTestRoot, []data.PublicKey{origRootKey}, 1, nil)
	require.NoError(t, err)
	prevRoot, err := data.RootFromSigned(signedOrigTestRoot)
	require.NoError(t, err)

	// TUF key with PEM-encoded x509 certificate
	replRootKey, err := testutils.CreateKey(cs, gun, data.CanonicalRootRole, keyAlg)
	require.NoError(t, err)

	rootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{replRootKey.ID()}, nil)
	require.NoError(t, err)

	testRoot, err := data.NewRoot(
		map[string]data.PublicKey{replRootKey.ID(): replRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &rootRole.RootRole,
			data.CanonicalTimestampRole: &rootRole.RootRole,
			data.CanonicalTargetsRole:   &rootRole.RootRole,
			data.CanonicalSnapshotRole:  &rootRole.RootRole},
		false,
	)
	testRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	signedTestRoot, err := testRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, signedTestRoot, []data.PublicKey{replRootKey, origRootKey}, 2, nil)
	require.NoError(t, err)

	typedSignedRoot, err := data.RootFromSigned(signedTestRoot)
	require.NoError(t, err)

	// This call to trustpinning.ValidateRoot will succeed since we are using a valid PEM
	// encoded certificate, and have no other certificates for this CN
	validatedRoot, err := trustpinning.ValidateRoot(prevRoot, signedTestRoot, gun, trustpinning.TrustPinConfig{})
	require.NoError(t, err)
	for idx, sig := range typedSignedRoot.Signatures {
		if sig.KeyID == replRootKey.ID() {
			typedSignedRoot.Signatures[idx].IsValid = true
		}
	}
	require.Equal(t, typedSignedRoot, validatedRoot)
}

// TestValidateRootRotationMissingOrigSig runs through a full root certificate rotation
// where we are missing the original root key signature. Verification should fail.
// We test this with both an RSA and ECDSA root certificate
func TestValidateRootRotationMissingOrigSig(t *testing.T) {
	testValidateRootRotationMissingOrigSig(t, data.ECDSAKey, data.ECDSAx509Key)
	if !testing.Short() {
		testValidateRootRotationMissingOrigSig(t, data.RSAKey, data.RSAx509Key)
	}
}

func testValidateRootRotationMissingOrigSig(t *testing.T, keyAlg, rootKeyType string) {
	var gun data.GUN = "docker.com/notary"

	memKeyStore := trustmanager.NewKeyMemoryStore(passphraseRetriever)
	cs := cryptoservice.NewCryptoService(memKeyStore)

	// TUF key with PEM-encoded x509 certificate
	origRootKey, err := testutils.CreateKey(cs, gun, data.CanonicalRootRole, keyAlg)
	require.NoError(t, err)

	origRootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{origRootKey.ID()}, nil)
	require.NoError(t, err)

	origTestRoot, err := data.NewRoot(
		map[string]data.PublicKey{origRootKey.ID(): origRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &origRootRole.RootRole,
			data.CanonicalTargetsRole:   &origRootRole.RootRole,
			data.CanonicalSnapshotRole:  &origRootRole.RootRole,
			data.CanonicalTimestampRole: &origRootRole.RootRole,
		},
		false,
	)
	origTestRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	signedOrigTestRoot, err := origTestRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, signedOrigTestRoot, []data.PublicKey{origRootKey}, 1, nil)
	require.NoError(t, err)
	prevRoot, err := data.RootFromSigned(signedOrigTestRoot)
	require.NoError(t, err)

	// TUF key with PEM-encoded x509 certificate
	replRootKey, err := testutils.CreateKey(cs, gun, data.CanonicalRootRole, keyAlg)
	require.NoError(t, err)

	rootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{replRootKey.ID()}, nil)
	require.NoError(t, err)

	testRoot, err := data.NewRoot(
		map[string]data.PublicKey{replRootKey.ID(): replRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &rootRole.RootRole,
			data.CanonicalTargetsRole:   &rootRole.RootRole,
			data.CanonicalSnapshotRole:  &rootRole.RootRole,
			data.CanonicalTimestampRole: &rootRole.RootRole,
		},
		false,
	)
	testRoot.Signed.Version = 2
	require.NoError(t, err, "Failed to create new root")

	signedTestRoot, err := testRoot.ToSigned()
	require.NoError(t, err)

	// We only sign with the new key, and not with the original one.
	err = signed.Sign(cs, signedTestRoot, []data.PublicKey{replRootKey}, 1, nil)
	require.NoError(t, err)

	// This call to trustpinning.ValidateRoot will fail since we don't have the original key's signature
	_, err = trustpinning.ValidateRoot(prevRoot, signedTestRoot, gun, trustpinning.TrustPinConfig{})
	require.Error(t, err, "insufficient signatures on root")

	// If we clear out an valid certs from the prevRoot, this will still fail
	prevRoot.Signed.Keys = nil
	_, err = trustpinning.ValidateRoot(prevRoot, signedTestRoot, gun, trustpinning.TrustPinConfig{})
	require.Error(t, err, "insufficient signatures on root")
}

// TestValidateRootRotationMissingNewSig runs through a full root certificate rotation
// where we are missing the new root key signature. Verification should fail.
// We test this with both an RSA and ECDSA root certificate
func TestValidateRootRotationMissingNewSig(t *testing.T) {
	testValidateRootRotationMissingNewSig(t, data.ECDSAKey, data.ECDSAx509Key)
	if !testing.Short() {
		testValidateRootRotationMissingNewSig(t, data.RSAKey, data.RSAx509Key)
	}
}

func testValidateRootRotationMissingNewSig(t *testing.T, keyAlg, rootKeyType string) {
	var gun data.GUN = "docker.com/notary"

	memKeyStore := trustmanager.NewKeyMemoryStore(passphraseRetriever)
	cs := cryptoservice.NewCryptoService(memKeyStore)

	// TUF key with PEM-encoded x509 certificate
	origRootKey, err := testutils.CreateKey(cs, gun, data.CanonicalRootRole, keyAlg)
	require.NoError(t, err)

	origRootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{origRootKey.ID()}, nil)
	require.NoError(t, err)

	origTestRoot, err := data.NewRoot(
		map[string]data.PublicKey{origRootKey.ID(): origRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &origRootRole.RootRole,
			data.CanonicalTargetsRole:   &origRootRole.RootRole,
			data.CanonicalSnapshotRole:  &origRootRole.RootRole,
			data.CanonicalTimestampRole: &origRootRole.RootRole,
		},
		false,
	)
	origTestRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	signedOrigTestRoot, err := origTestRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, signedOrigTestRoot, []data.PublicKey{origRootKey}, 1, nil)
	require.NoError(t, err)
	prevRoot, err := data.RootFromSigned(signedOrigTestRoot)
	require.NoError(t, err)

	// TUF key with PEM-encoded x509 certificate
	replRootKey, err := testutils.CreateKey(cs, gun, data.CanonicalRootRole, keyAlg)
	require.NoError(t, err)

	rootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{replRootKey.ID()}, nil)
	require.NoError(t, err)

	testRoot, err := data.NewRoot(
		map[string]data.PublicKey{replRootKey.ID(): replRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &rootRole.RootRole,
			data.CanonicalTargetsRole:   &rootRole.RootRole,
			data.CanonicalSnapshotRole:  &rootRole.RootRole,
			data.CanonicalTimestampRole: &rootRole.RootRole,
		},
		false,
	)
	require.NoError(t, err, "Failed to create new root")

	signedTestRoot, err := testRoot.ToSigned()
	require.NoError(t, err)

	// We only sign with the old key, and not with the new one
	err = signed.Sign(cs, signedTestRoot, []data.PublicKey{origRootKey}, 1, nil)
	require.NoError(t, err)

	// This call to trustpinning.ValidateRoot will succeed since we are using a valid PEM
	// encoded certificate, and have no other certificates for this CN
	_, err = trustpinning.ValidateRoot(prevRoot, signedTestRoot, gun, trustpinning.TrustPinConfig{})
	require.Error(t, err, "insufficient signatures on root")
}

// TestValidateRootRotationTrustPinning runs a full root certificate rotation but ensures that
// the specified trust pinning is respected with the new root for the Certs and TOFUs settings
func TestValidateRootRotationTrustPinning(t *testing.T) {
	// The gun to test
	var gun data.GUN = "docker.com/notary"

	memKeyStore := trustmanager.NewKeyMemoryStore(passphraseRetriever)
	cs := cryptoservice.NewCryptoService(memKeyStore)

	// TUF key with PEM-encoded x509 certificate
	origRootKey, err := testutils.CreateKey(cs, gun, data.CanonicalRootRole, data.RSAKey)
	require.NoError(t, err)

	origRootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{origRootKey.ID()}, nil)
	require.NoError(t, err)

	origTestRoot, err := data.NewRoot(
		map[string]data.PublicKey{origRootKey.ID(): origRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &origRootRole.RootRole,
			data.CanonicalTargetsRole:   &origRootRole.RootRole,
			data.CanonicalSnapshotRole:  &origRootRole.RootRole,
			data.CanonicalTimestampRole: &origRootRole.RootRole,
		},
		false,
	)
	origTestRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	signedOrigTestRoot, err := origTestRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, signedOrigTestRoot, []data.PublicKey{origRootKey}, 1, nil)
	require.NoError(t, err)
	prevRoot, err := data.RootFromSigned(signedOrigTestRoot)
	require.NoError(t, err)

	// TUF key with PEM-encoded x509 certificate
	replRootKey, err := testutils.CreateKey(cs, gun, data.CanonicalRootRole, data.RSAKey)
	require.NoError(t, err)

	rootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{replRootKey.ID()}, nil)
	require.NoError(t, err)

	testRoot, err := data.NewRoot(
		map[string]data.PublicKey{replRootKey.ID(): replRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &rootRole.RootRole,
			data.CanonicalTimestampRole: &rootRole.RootRole,
			data.CanonicalTargetsRole:   &rootRole.RootRole,
			data.CanonicalSnapshotRole:  &rootRole.RootRole},
		false,
	)
	testRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	signedTestRoot, err := testRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, signedTestRoot, []data.PublicKey{replRootKey, origRootKey}, 2, nil)
	require.NoError(t, err)

	typedSignedRoot, err := data.RootFromSigned(signedTestRoot)
	require.NoError(t, err)

	// This call to trustpinning.ValidateRoot will fail due to the trust pinning mismatch in certs
	invalidCertConfig := trustpinning.TrustPinConfig{
		Certs: map[string][]string{
			gun.String(): {origRootKey.ID()},
		},
		DisableTOFU: true,
	}
	_, err = trustpinning.ValidateRoot(prevRoot, signedTestRoot, gun, invalidCertConfig)
	require.Error(t, err)

	// This call will succeed since we include the new root cert ID (and the old one)
	validCertConfig := trustpinning.TrustPinConfig{
		Certs: map[string][]string{
			gun.String(): {origRootKey.ID(), replRootKey.ID()},
		},
		DisableTOFU: true,
	}

	validatedRoot, err := trustpinning.ValidateRoot(prevRoot, signedTestRoot, gun, validCertConfig)
	require.NoError(t, err)
	for idx, sig := range typedSignedRoot.Signatures {
		if sig.KeyID == replRootKey.ID() {
			typedSignedRoot.Signatures[idx].IsValid = true
		}
	}
	require.Equal(t, typedSignedRoot, validatedRoot)

	// This call will also succeed since we only need the new replacement root ID to be pinned
	validCertConfig = trustpinning.TrustPinConfig{
		Certs: map[string][]string{
			gun.String(): {replRootKey.ID()},
		},
		DisableTOFU: true,
	}
	validatedRoot, err = trustpinning.ValidateRoot(prevRoot, signedTestRoot, gun, validCertConfig)
	require.NoError(t, err)
	require.Equal(t, typedSignedRoot, validatedRoot)

	// Even if we disable TOFU in the trustpinning, since we have a previously trusted root we should honor a valid rotation
	validatedRoot, err = trustpinning.ValidateRoot(prevRoot, signedTestRoot, gun, trustpinning.TrustPinConfig{DisableTOFU: true})
	require.NoError(t, err)
	require.Equal(t, typedSignedRoot, validatedRoot)
}

// TestValidateRootRotationTrustPinningInvalidCA runs a full root certificate rotation but ensures that
// the specified trust pinning rejects the new root for not being signed by the specified CA
func TestValidateRootRotationTrustPinningInvalidCA(t *testing.T) {
	var gun data.GUN = "notary-signer"
	keyAlg := data.RSAKey
	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	defer os.RemoveAll(tempBaseDir)
	require.NoError(t, err, "failed to create a temporary directory: %s", err)

	leafCert, err := utils.LoadCertFromFile("../fixtures/notary-signer.crt")
	require.NoError(t, err)

	intermediateCert, err := utils.LoadCertFromFile("../fixtures/intermediate-ca.crt")
	require.NoError(t, err)

	pemChainBytes, err := utils.CertChainToPEM([]*x509.Certificate{leafCert, intermediateCert})
	require.NoError(t, err)

	origRootKey := data.NewPublicKey(data.RSAx509Key, pemChainBytes)

	rootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{origRootKey.ID()}, nil)
	require.NoError(t, err)

	testRoot, err := data.NewRoot(
		map[string]data.PublicKey{origRootKey.ID(): origRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &rootRole.RootRole,
			data.CanonicalTimestampRole: &rootRole.RootRole,
			data.CanonicalTargetsRole:   &rootRole.RootRole,
			data.CanonicalSnapshotRole:  &rootRole.RootRole},
		false,
	)
	testRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	pemBytes, err := ioutil.ReadFile("../fixtures/notary-signer.key")
	require.NoError(t, err, "could not read key file")
	privKey, err := utils.ParsePEMPrivateKey(pemBytes, "")
	require.NoError(t, err)

	store, err := trustmanager.NewKeyFileStore(tempBaseDir, passphraseRetriever)
	require.NoError(t, err)
	cs := cryptoservice.NewCryptoService(store)

	err = store.AddKey(trustmanager.KeyInfo{Role: data.CanonicalRootRole, Gun: gun}, privKey)
	require.NoError(t, err)

	origSignedTestRoot, err := testRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, origSignedTestRoot, []data.PublicKey{origRootKey}, 1, nil)
	require.NoError(t, err)
	prevRoot, err := data.RootFromSigned(origSignedTestRoot)
	require.NoError(t, err)

	// generate a new TUF key with PEM-encoded x509 certificate, not signed by our pinned CA
	replRootKey, err := testutils.CreateKey(cs, gun, data.CanonicalRootRole, keyAlg)
	require.NoError(t, err)

	_, err = data.NewRole(data.CanonicalRootRole, 1, []string{replRootKey.ID()}, nil)
	require.NoError(t, err)
	newRoot, err := data.NewRoot(
		map[string]data.PublicKey{replRootKey.ID(): replRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &rootRole.RootRole,
			data.CanonicalTimestampRole: &rootRole.RootRole,
			data.CanonicalTargetsRole:   &rootRole.RootRole,
			data.CanonicalSnapshotRole:  &rootRole.RootRole},
		false,
	)
	newRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	newSignedTestRoot, err := newRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, newSignedTestRoot, []data.PublicKey{replRootKey, origRootKey}, 2, nil)
	require.NoError(t, err)

	// Check that we respect the trust pinning on rotation
	validCAFilepath := "../fixtures/root-ca.crt"
	_, err = trustpinning.ValidateRoot(prevRoot, newSignedTestRoot, gun, trustpinning.TrustPinConfig{CA: map[string]string{gun.String(): validCAFilepath}, DisableTOFU: true})
	require.Error(t, err)
}

func generateTestingCertificate(rootKey data.PrivateKey, gun data.GUN, timeToExpire time.Duration) (*x509.Certificate, error) {
	startTime := time.Now()
	return cryptoservice.GenerateCertificate(rootKey, gun, startTime, startTime.Add(timeToExpire))
}

func generateExpiredTestingCertificate(rootKey data.PrivateKey, gun data.GUN) (*x509.Certificate, error) {
	startTime := time.Now().AddDate(-10, 0, 0)
	return cryptoservice.GenerateCertificate(rootKey, gun, startTime, startTime.AddDate(1, 0, 0))
}

func TestParsePEMPublicKey(t *testing.T) {
	var gun data.GUN = "notary"
	memStore := trustmanager.NewKeyMemoryStore(passphraseRetriever)
	cs := cryptoservice.NewCryptoService(memStore)

	// can parse ECDSA PEM
	ecdsaPubKey, err := cs.Create("root", "docker.io/notary/test", data.ECDSAKey)
	require.NoError(t, err)
	ecdsaPemBytes := pem.EncodeToMemory(&pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   ecdsaPubKey.Public(),
	})

	ecdsaParsedPubKey, err := utils.ParsePEMPublicKey(ecdsaPemBytes)
	require.NoError(t, err, "no key: %s", ecdsaParsedPubKey.Public())

	// can parse certificates
	ecdsaPrivKey, _, err := memStore.GetKey(ecdsaPubKey.ID())
	require.NoError(t, err)
	cert, err := generateTestingCertificate(ecdsaPrivKey, gun, notary.Day*30)
	require.NoError(t, err)
	ecdsaPubKeyFromCert, err := utils.ParsePEMPublicKey(utils.CertToPEM(cert))
	require.NoError(t, err)

	thatData := []byte{1, 2, 3, 4}
	sig, err := ecdsaPrivKey.Sign(rand.Reader, thatData, nil)
	require.NoError(t, err)
	err = signed.ECDSAVerifier{}.Verify(ecdsaPubKeyFromCert, sig, thatData)
	require.NoError(t, err)

	// can parse RSA PEM
	rsaPubKey, err := cs.Create("root", "docker.io/notary/test2", data.RSAKey)
	require.NoError(t, err)
	rsaPemBytes := pem.EncodeToMemory(&pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   rsaPubKey.Public(),
	})
	_, err = utils.ParsePEMPublicKey(rsaPemBytes)
	require.NoError(t, err)

	// unsupported key type
	unsupportedPemBytes := pem.EncodeToMemory(&pem.Block{
		Type:    "PRIVATE KEY",
		Headers: nil,
		Bytes:   []byte{0, 0, 0, 0},
	})
	_, err = utils.ParsePEMPublicKey(unsupportedPemBytes)
	require.Error(t, err)

	// bad key
	badPemBytes := pem.EncodeToMemory(&pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   []byte{0, 0, 0, 0},
	})
	_, err = utils.ParsePEMPublicKey(badPemBytes)
	require.Error(t, err)
}

func TestCheckingCertExpiry(t *testing.T) {
	var gun data.GUN = "notary"
	memStore := trustmanager.NewKeyMemoryStore(passphraseRetriever)
	cs := cryptoservice.NewCryptoService(memStore)
	testPubKey, err := cs.Create(data.CanonicalRootRole, gun, data.ECDSAKey)
	require.NoError(t, err)
	testPrivKey, _, err := memStore.GetKey(testPubKey.ID())
	require.NoError(t, err)

	almostExpiredCert, err := generateTestingCertificate(testPrivKey, gun, notary.Day*30)
	require.NoError(t, err)
	almostExpiredPubKey, err := utils.ParsePEMPublicKey(utils.CertToPEM(almostExpiredCert))
	require.NoError(t, err)

	// set up a logrus logger to capture warning output
	origLevel := logrus.GetLevel()
	logrus.SetLevel(logrus.WarnLevel)
	defer logrus.SetLevel(origLevel)
	logBuf := bytes.NewBuffer(nil)
	logrus.SetOutput(logBuf)

	rootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{almostExpiredPubKey.ID()}, nil)
	require.NoError(t, err)
	testRoot, err := data.NewRoot(
		map[string]data.PublicKey{almostExpiredPubKey.ID(): almostExpiredPubKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &rootRole.RootRole,
			data.CanonicalTimestampRole: &rootRole.RootRole,
			data.CanonicalTargetsRole:   &rootRole.RootRole,
			data.CanonicalSnapshotRole:  &rootRole.RootRole},
		false,
	)
	testRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	signedTestRoot, err := testRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, signedTestRoot, []data.PublicKey{almostExpiredPubKey}, 1, nil)
	require.NoError(t, err)

	// This is a valid root certificate, but check that we get a Warn-level message that the certificate is near expiry
	_, err = trustpinning.ValidateRoot(nil, signedTestRoot, gun, trustpinning.TrustPinConfig{})
	require.NoError(t, err)
	require.Contains(t, logBuf.String(), fmt.Sprintf("certificate with CN %s is near expiry", gun))

	expiredCert, err := generateExpiredTestingCertificate(testPrivKey, gun)
	require.NoError(t, err)
	expiredPubKey := utils.CertToKey(expiredCert)

	rootRole, err = data.NewRole(data.CanonicalRootRole, 1, []string{expiredPubKey.ID()}, nil)
	require.NoError(t, err)
	testRoot, err = data.NewRoot(
		map[string]data.PublicKey{expiredPubKey.ID(): expiredPubKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &rootRole.RootRole,
			data.CanonicalTimestampRole: &rootRole.RootRole,
			data.CanonicalTargetsRole:   &rootRole.RootRole,
			data.CanonicalSnapshotRole:  &rootRole.RootRole},
		false,
	)
	testRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	signedTestRoot, err = testRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, signedTestRoot, []data.PublicKey{expiredPubKey}, 1, nil)
	require.NoError(t, err)

	// This is an invalid root certificate since it's expired
	_, err = trustpinning.ValidateRoot(nil, signedTestRoot, gun, trustpinning.TrustPinConfig{})
	require.Error(t, err)
}

func TestValidateRootWithExpiredIntermediate(t *testing.T) {
	now := time.Now()
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	memStore := trustmanager.NewKeyMemoryStore(passphraseRetriever)
	cs := cryptoservice.NewCryptoService(memStore)

	// generate CA cert
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	require.NoError(t, err)
	caTmpl := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "notary testing CA",
		},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:       true,
		MaxPathLen: 3,
	}
	caPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	_, err = x509.CreateCertificate(
		rand.Reader,
		&caTmpl,
		&caTmpl,
		caPrivKey.Public(),
		caPrivKey,
	)
	require.NoError(t, err)

	// generate expired intermediate
	intTmpl := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "EXPIRED notary testing intermediate",
		},
		NotBefore:             now.Add(-2 * notary.Year),
		NotAfter:              now.Add(-notary.Year),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:       true,
		MaxPathLen: 2,
	}
	intPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	intCert, err := x509.CreateCertificate(
		rand.Reader,
		&intTmpl,
		&caTmpl,
		intPrivKey.Public(),
		caPrivKey,
	)
	require.NoError(t, err)

	// generate leaf
	serialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
	require.NoError(t, err)
	leafTmpl := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "docker.io/notary/test",
		},
		NotBefore: now.Add(-time.Hour),
		NotAfter:  now.Add(time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
		BasicConstraintsValid: true,
	}

	leafPubKey, err := cs.Create("root", "docker.io/notary/test", data.ECDSAKey)
	require.NoError(t, err)
	leafPrivKey, _, err := cs.GetPrivateKey(leafPubKey.ID())
	require.NoError(t, err)
	signer := leafPrivKey.CryptoSigner()
	leafCert, err := x509.CreateCertificate(
		rand.Reader,
		&leafTmpl,
		&intTmpl,
		signer.Public(),
		intPrivKey,
	)
	require.NoError(t, err)

	rootBundleWriter := bytes.NewBuffer(nil)
	pem.Encode(
		rootBundleWriter,
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: leafCert,
		},
	)
	pem.Encode(
		rootBundleWriter,
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: intCert,
		},
	)

	rootBundle := rootBundleWriter.Bytes()

	ecdsax509Key := data.NewECDSAx509PublicKey(rootBundle)

	otherKey, err := cs.Create("targets", "docker.io/notary/test", data.ED25519Key)
	require.NoError(t, err)

	root := data.SignedRoot{
		Signatures: make([]data.Signature, 0),
		Signed: data.Root{
			SignedCommon: data.SignedCommon{
				Type:    "Root",
				Expires: now.Add(time.Hour),
				Version: 1,
			},
			Keys: map[string]data.PublicKey{
				ecdsax509Key.ID(): ecdsax509Key,
				otherKey.ID():     otherKey,
			},
			Roles: map[data.RoleName]*data.RootRole{
				"root": {
					KeyIDs:    []string{ecdsax509Key.ID()},
					Threshold: 1,
				},
				"targets": {
					KeyIDs:    []string{otherKey.ID()},
					Threshold: 1,
				},
				"snapshot": {
					KeyIDs:    []string{otherKey.ID()},
					Threshold: 1,
				},
				"timestamp": {
					KeyIDs:    []string{otherKey.ID()},
					Threshold: 1,
				},
			},
		},
		Dirty: true,
	}

	signedRoot, err := root.ToSigned()
	require.NoError(t, err)
	err = signed.Sign(cs, signedRoot, []data.PublicKey{ecdsax509Key}, 1, nil)
	require.NoError(t, err)

	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	defer os.RemoveAll(tempBaseDir)
	require.NoError(t, err, "failed to create a temporary directory: %s", err)

	_, err = trustpinning.ValidateRoot(
		nil,
		signedRoot,
		"docker.io/notary/test",
		trustpinning.TrustPinConfig{},
	)
	require.Error(t, err, "failed to invalidate expired intermediate certificate")
}

func TestCheckingWildcardCert(t *testing.T) {
	memStore := trustmanager.NewKeyMemoryStore(passphraseRetriever)
	cs := cryptoservice.NewCryptoService(memStore)
	testPubKey, err := cs.Create(data.CanonicalRootRole, "docker.io/notary/*", data.ECDSAKey)
	require.NoError(t, err)
	testPrivKey, _, err := memStore.GetKey(testPubKey.ID())
	require.NoError(t, err)

	testCert, err := generateTestingCertificate(testPrivKey, "docker.io/notary/*", notary.Year)
	require.NoError(t, err)
	testCertPubKey, err := utils.ParsePEMPublicKey(utils.CertToPEM(testCert))
	require.NoError(t, err)

	rootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{testCertPubKey.ID()}, nil)
	require.NoError(t, err)
	testRoot, err := data.NewRoot(
		map[string]data.PublicKey{testCertPubKey.ID(): testCertPubKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &rootRole.RootRole,
			data.CanonicalTimestampRole: &rootRole.RootRole,
			data.CanonicalTargetsRole:   &rootRole.RootRole,
			data.CanonicalSnapshotRole:  &rootRole.RootRole},
		false,
	)
	testRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	signedTestRoot, err := testRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, signedTestRoot, []data.PublicKey{testCertPubKey}, 1, nil)
	require.NoError(t, err)

	_, err = trustpinning.ValidateRoot(
		nil,
		signedTestRoot,
		"docker.io/notary/test",
		trustpinning.TrustPinConfig{},
	)
	require.NoError(t, err, "expected wildcard cert to validate")

	_, err = trustpinning.ValidateRoot(
		nil,
		signedTestRoot,
		"docker.io/not-a-match",
		trustpinning.TrustPinConfig{},
	)
	require.Error(t, err, "expected wildcard cert not to validate")
}

func TestWildcardMatching(t *testing.T) {
	var wildcardTests = []struct {
		CN  string
		gun string
		out bool
	}{
		{"docker.com/*", "docker.com/notary", true},
		{"docker.com/**", "docker.com/notary", true},
		{"*", "docker.com/any", true},
		{"*", "", true},
		{"**", "docker.com/any", true},
		{"test/*******", "test/many/wildcard", true},
		{"test/**/*/", "test/test", false},
		{"test/*/wild", "test/test/wild", false},
		{"*/all", "test/all", false},
		{"docker.com/*/*", "docker.com/notary/test", false},
		{"docker.com/*/**", "docker.com/notary/test", false},
		{"", "*", false},
		{"*abc*", "abc", false},
		{"test/*/wild*", "test/test/wild", false},
	}
	for _, tt := range wildcardTests {
		require.Equal(t, trustpinning.MatchCNToGun(tt.CN, data.GUN(tt.gun)), tt.out)
	}
}
