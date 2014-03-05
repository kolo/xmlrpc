package xmlrpc

import (
	"testing"
)

const ARRAY_VALUE = `
<value>
	<array>
		<data>
			<value><int>1</int></value>
			<value><int>5</int></value>
			<value><int>13</int></value>
		</data>
	</array>
</value>
`

func Test_unmarshalArray(t *testing.T) {
	s := make([]int, 0)
	if err := unmarshal([]byte(ARRAY_VALUE), &s); err != nil {
		t.Fatal(err)
	}

	t.Log(s)
}
