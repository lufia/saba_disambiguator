package sabadisambiguator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type PerceptronClassifier struct {
	Weight    map[string]float64
	CumWeight map[string]float64
	Count     int
}

func newPerceptronClassifier() *PerceptronClassifier {
	return &PerceptronClassifier{make(map[string]float64), make(map[string]float64), 1}
}

func NewPerceptronClassifier(examples Examples) *PerceptronClassifier {
	train_, dev := splitTrainAndDev(examples)
	train := overSampling(train_)
	model := newPerceptronClassifier()
	for iter := 0; iter < 30; iter++ {
		shuffle(train)
		for _, example := range train {
			model.learn(*example)
		}

		devPredicts := make([]LabelType, len(dev))
		for i, example := range dev {
			devPredicts[i] = model.Predict(example.Fv)
		}
		accuracy := GetAccuracy(ExtractGoldLabels(dev), devPredicts)
		precision := GetPrecision(ExtractGoldLabels(dev), devPredicts)
		recall := GetRecall(ExtractGoldLabels(dev), devPredicts)
		f := (2 * recall * precision) / (recall + precision)
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Iter:%d\tAccuracy:%0.03f\tPrecision:%0.03f\tRecall:%0.03f\tF-value:%0.03f", iter, accuracy, precision, recall, f))
	}
	for _, e := range dev {
		predLabel := model.Predict(e.Fv)
		if predLabel != e.Label {
			url := fmt.Sprintf("https://twitter.com/%s/status/%s", e.Tweet.User.UserName, e.Tweet.ID)
			fmt.Println(fmt.Sprintf("%d\t%d\t%s", e.Label, predLabel, url))
		}
	}

	return model
}

func (model *PerceptronClassifier) learn(example Example) {
	predict := model.predictForTraining(example.Fv)
	if example.Label != predict {
		for _, f := range example.Fv {
			w, _ := model.Weight[f]
			cumW, _ := model.CumWeight[f]
			model.Weight[f] = w + float64(example.Label)*1.0
			model.CumWeight[f] = cumW + float64(model.Count)*float64(example.Label)*1.0
		}
		model.Count += 1
	}
}

func (model *PerceptronClassifier) predictForTraining(features FeatureVector) LabelType {
	result := 0.0
	for _, f := range features {
		w, ok := model.Weight[f]
		if ok {
			result = result + w*1.0
		}
	}
	if result > 0 {
		return POSITIVE
	}
	return NEGATIVE
}

func (model PerceptronClassifier) PredictScore(features FeatureVector) float64 {
	result := 0.0
	for _, f := range features {
		w, ok := model.Weight[f]
		if ok {
			result = result + w*1.0
		}

		w, ok = model.CumWeight[f]
		if ok {
			result = result - w*1.0/float64(model.Count)
		}

	}
	return result
}

func (model PerceptronClassifier) Predict(features FeatureVector) LabelType {
	if model.PredictScore(features) > 0 {
		return POSITIVE
	}
	return NEGATIVE
}

func ExtractGoldLabels(examples Examples) []LabelType {
	golds := make([]LabelType, 0, 0)
	for _, e := range examples {
		golds = append(golds, e.Label)
	}
	return golds
}

func WritePerceptron(perceptron PerceptronClassifier, filename string) error {
	perceptronJson, err := json.Marshal(perceptron)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, perceptronJson, 0644)
	if err != nil {
		return err
	}

	return nil
}

func LoadPerceptron(filename string) (*PerceptronClassifier, error) {
	perceptron := PerceptronClassifier{}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &perceptron); err != nil {
		return nil, err
	}
	return &perceptron, nil
}
