package matchers

import "bytes"

// Glb matches a glTF model format file.
func Glb(in []byte, _ uint32) bool {
	// GLB is the binary file format representation of 3D models save in
	// the GL transmission Format (glTF).
	// see more: https://docs.fileformat.com/3d/glb/
	//           https://www.iana.org/assignments/media-types/model/gltf-binary
	if len(in) < 8 {
		return false
	}

	// GLB file format is based on little endian and its header structure
	// show  below:
	//
	// <-- 12-byte header                             -->
	// | magic            | version          | length   |
	// | (uint32)         | (uint32)         | (uint32) |
	// | \x67\x6C\x54\x46 | \x01\x00\x00\x00 | ...      |
	// | g   l   T   F    | 1                | ...      |
	return bytes.HasPrefix(in, []byte("\x67\x6C\x54\x46\x02\x00\x00\x00")) ||
		bytes.HasPrefix(in, []byte("\x67\x6C\x54\x46\x01\x00\x00\x00"))
}
