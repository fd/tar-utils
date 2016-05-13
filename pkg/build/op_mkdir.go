package tarbuild

func applyMKDIR(dst, src *Dir, op tarOp) error {
	for _, n := range op.Args {
		_, err := dst.MkdirAll(n)
		if err != nil {
			return err
		}
	}
	dst.BakeDeepEntries()
	return nil
}
