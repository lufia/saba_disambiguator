package sabadisambiguator

func splitTrainAndDev(examples Examples) (train Examples, dev Examples) {
	shuffle(examples)
	n := int(0.8 * float64(len(examples)))
	return examples[0:n], examples[n:]
}
