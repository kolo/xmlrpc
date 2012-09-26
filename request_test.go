package xmlrpc

import (
    "testing"
)

func Test_buildValueElement_EscapesString(t *testing.T) {
    escaped := buildValueElement("Johnson & Johnson > 1")
    assert_equal(t, "<value><string>Johnson &amp; Johnson &gt; 1</string></value>", escaped)
}
