package sabadisambiguator

import "math"

func splitTrainAndDev(examples Examples) (train Examples, dev Examples) {
	shuffle(examples)
	n := int(0.8 * float64(len(examples)))
	return examples[0:n], examples[n:]
}

func overSampling(examples Examples) Examples {
	result := examples
	positiveExamples := Examples{}
	negativeExamples := Examples{}

	for _, e := range examples {
		if e.Label == POSITIVE {
			positiveExamples = append(positiveExamples, e)
		} else if e.Label == NEGATIVE {
			negativeExamples = append(negativeExamples, e)
		}
	}
	n := len(negativeExamples) - len(positiveExamples)
	examplesToBeOverSampled := Examples{}
	if n > 0 {
		examplesToBeOverSampled = positiveExamples
	} else {
		examplesToBeOverSampled = negativeExamples
	}

	for i := 0; i < int(math.Abs(float64(n))); i++ {
		shuffle(examplesToBeOverSampled)
		result = append(result, examplesToBeOverSampled[0])
	}
	shuffle(result)
	return result
}
