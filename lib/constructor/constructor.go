package constructor

import "github.com/Laughs-In-Flowers/countfloyd/lib/feature"

type Registerable func() feature.Constructor

type Registerables struct {
	has []Registerable
}

func (r *Registerables) Register() error {
	var err error
	for _, v := range r.has {
		err = feature.SetConstructor(v())
		if err != nil {
			return err
		}
	}
	return nil
}

var register *Registerables

func init() {
	register = &Registerables{
		[]Registerable{
			CollectionMember,
			CombinationStrings,
			List,
			ListWithNull,
			ListShuffle,
			ListExpandIntRange,
			ListMirrorInts,
			Set,
			SimpleRandom,
			SourcedRandom,
			WeightedStringWithWeights,
			WeightedStringWithNormalizedWeights,
		},
	}

	register.Register()
}
