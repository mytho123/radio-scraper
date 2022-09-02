package sources

import "time"

type Scrap func(start, end time.Time, results chan<- Track) error

type Track struct {
	Artist string
	Title  string
}

func GetSources() map[string]Scrap {
	return map[string]Scrap{
		"ouifm": func(start, end time.Time, results chan<- Track) error {
			return scrapGroupe1981("2174546520932614531", start, end, results)
		},
		"ouifm_rock_inde": func(start, end time.Time, results chan<- Track) error {
			return scrapGroupe1981("3134161803443976526", start, end, results)
		},
		"ouifm_bring_the_noise": func(start, end time.Time, results chan<- Track) error {
			return scrapGroupe1981("4004502594738215513", start, end, results)
		},
		"ouifm_reggae": func(start, end time.Time, results chan<- Track) error {
			return scrapGroupe1981("3540892623380233022", start, end, results)
		},
		"ouifm_accoustic": func(start, end time.Time, results chan<- Track) error {
			return scrapGroupe1981("3906034555622012146", start, end, results)
		},
		"latina": func(start, end time.Time, results chan<- Track) error {
			return scrapGroupe1981("2174546520932614634", start, end, results)
		},
		"voltage": func(start, end time.Time, results chan<- Track) error {
			return scrapGroupe1981("2174546520932614870", start, end, results)
		},
		"virgin_radio": func(start, end time.Time, results chan<- Track) error {
			return scrapRfm("https://direct-radio.rfm.fr/playout?radio=2&limit=20", start, end, results)
		},
	}
}
