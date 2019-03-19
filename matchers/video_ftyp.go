package matchers

var (
	mp4Sigs = []sig{
		ftypSig("avc1"), ftypSig("dash"), ftypSig("iso2"), ftypSig("iso3"),
		ftypSig("iso4"), ftypSig("iso5"), ftypSig("iso6"), ftypSig("isom"),
		ftypSig("mmp4"), ftypSig("mp41"), ftypSig("mp42"), ftypSig("mp4v"),
		ftypSig("mp71"), ftypSig("MSNV"), ftypSig("NDAS"), ftypSig("NDSC"),
		ftypSig("NSDC"), ftypSig("NSDH"), ftypSig("NDSM"), ftypSig("NDSP"),
		ftypSig("NDSS"), ftypSig("NDXC"), ftypSig("NDXH"), ftypSig("NDXM"),
		ftypSig("NDXP"), ftypSig("NDXS"), ftypSig("F4V "), ftypSig("F4P "),
	}
	threeGPSigs = []sig{
		ftypSig("3gp1"), ftypSig("3gp2"), ftypSig("3gp3"), ftypSig("3gp4"),
		ftypSig("3gp5"), ftypSig("3gp6"), ftypSig("3gs7"), ftypSig("3ge6"),
		ftypSig("3ge7"), ftypSig("3gg6"),
	}
	threeG2Sigs = []sig{
		ftypSig("3g2a"), ftypSig("3g2b"), ftypSig("3g2c"), ftypSig("KDDI"),
	}
	// TODO: add support for remaining video formats at ftyps.com
)

// Mp4 matches an MP4 file.
func Mp4(in []byte) bool {
	return detect(in, mp4Sigs)
}

// ThreeGP matches a 3GPP file.
func ThreeGP(in []byte) bool {
	return detect(in, threeGPSigs)
}

// ThreeG2 matches a 3GPP2 file.
func ThreeG2(in []byte) bool {
	return detect(in, threeG2Sigs)
}
