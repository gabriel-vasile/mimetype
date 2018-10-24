package mimetype

import (
	"testing"

	"github.com/gabriel-vasile/mimetype/matchers"
)

func TestAddNodeNeedingFullData(t *testing.T) {
	testNode1 := NewNode("test/node1; charset=utf-8", "test", matchers.False,
		Svg, X3d, Kml, Collada, Gml, Gpx)
	testNode2 := NewNode("test/node2; charset=utf-8", "test", matchers.False,
		Xlsx, Docx, Pptx, Epub, Jar)
	defer func() { //clearing global map after test is executed
		delete(FullDataNodesMap, testNode1.Mime())
		delete(FullDataNodesMap, testNode2.Mime())
	}()
	type args struct {
		root  *Node
		child []*Node
	}
	tests := []struct {
		name        string
		args        args
		expectedLen int
	}{
		{name: "add1node",
			args:        args{root: testNode1, child: []*Node{Svg}},
			expectedLen: 1,
		},
		{name: "add2nodes",
			args:        args{root: testNode1, child: []*Node{Gml}},
			expectedLen: 2,
		},
		{name: "add3NodesAtATime",
			args:        args{root: testNode2, child: []*Node{Xlsx, Docx, Pptx}},
			expectedLen: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddNodeNeedingFullData(tt.args.root, tt.args.child...)
			if len(FullDataNodesMap[tt.args.root.Mime()].children) != tt.expectedLen {
				t.Errorf("expected map-length=%d, got map-length=%d", tt.expectedLen, len(FullDataNodesMap[tt.args.root.Mime()].children))
			}
		})
	}
}
