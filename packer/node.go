// Copyright © 2019 Master.G
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package packer

// Node represents a tree node of the packer
type Node struct {
	child []*Node
	rc    Rectangle
	image *ImageInfo
}

func (n *Node) isLeaf() bool {
	if n == nil {
		return false
	}
	return len(n.child) == 0
}

func (n *Node) insert(img *ImageInfo) *Node {

	if !n.isLeaf() {
		// try inserting into first child
		newNode := n.child[0].insert(img)
		if newNode != nil {
			return newNode
		}

		// no room, insert into second
		return n.child[1].insert(img)
	} else {
		// if already contains a texture, return
		if n.image != nil {
			return nil
		}

		if n.rc.Width() < img.PaddedWidth() || n.rc.Height() < img.PaddedHeight() {
			// if too small, return
			return nil
		}

		if n.rc.Width() == img.PaddedWidth() && n.rc.Height() == img.PaddedHeight() {
			// PERFECT FIT, WOW
			return n
		}

		// split this node
		n.child = make([]*Node, 2)

		// decide which direction to split
		binWidth := img.PaddedWidth()
		binHeight := img.PaddedHeight()
		dw := n.rc.Width() - binWidth
		dh := n.rc.Height() - binHeight

		if dw > dh {
			n.child[0] = &Node{
				rc: Rectangle{
					Left:   n.rc.Left,
					Top:    n.rc.Top,
					Right:  n.rc.Left + binWidth,
					Bottom: n.rc.Bottom,
				},
			}
			n.child[1] = &Node{
				rc: Rectangle{
					Left:   n.rc.Left + binWidth,
					Top:    n.rc.Top,
					Right:  n.rc.Right,
					Bottom: n.rc.Bottom,
				},
			}
		} else {
			n.child[0] = &Node{
				rc: Rectangle{
					Left:   n.rc.Left,
					Top:    n.rc.Top,
					Right:  n.rc.Right,
					Bottom: n.rc.Top + binHeight,
				},
			}
			n.child[1] = &Node{
				rc: Rectangle{
					Left:   n.rc.Left,
					Top:    n.rc.Top + binHeight,
					Right:  n.rc.Right,
					Bottom: n.rc.Bottom,
				},
			}
		}

		// insert into the first child we just created
		return n.child[0].insert(img)
	}
}
