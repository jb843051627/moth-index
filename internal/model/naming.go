package model

import "strings"

func NormalizeHabitat(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(value)), " ")
}

func NormalizeKind(value string) string { return strings.TrimSpace(strings.ToLower(value)) }

func IsKnownTrapKind(value string) bool {
	value = NormalizeKind(value)
	for _, candidate := range ValidTrapKinds() {
		if value == candidate {
			return true
		}
	}
	return false
}

func TaxonKey(family, genus, species string) string {
	return strings.ToLower(strings.Join([]string{strings.TrimSpace(family), strings.TrimSpace(genus), strings.TrimSpace(species)}, ":"))
}

func SameTaxon(left, right Taxon) bool {
	return TaxonKey(left.Family, left.Genus, left.Species) == TaxonKey(right.Family, right.Genus, right.Species)
}

func MergeTags(values []Specimen) string {
	tags := make([]string, 0, len(values))
	for _, value := range values {
		tag := NormalizeTag(value.Tag)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return strings.Join(tags, ",")
}
