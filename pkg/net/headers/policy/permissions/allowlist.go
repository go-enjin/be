// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package permissions

type AllowList []Origin

func (a AllowList) IsEmpty() (empty bool) {
	empty = len(a) == 0
	return
}

func (a AllowList) HasNone() (present bool) {
	for _, o := range a {
		if present = o.OriginType() == AllowNone.OriginType() && o.Value() == AllowNone.Value(); present {
			break
		}
	}
	return
}

func (a AllowList) HasAll() (present bool) {
	for _, o := range a {
		if present = o.OriginType() == AllowAll.OriginType() && o.Value() == AllowAll.Value(); present {
			break
		}
	}
	return
}

func (a AllowList) Value() (value string) {
	if a.HasNone() {
		value = AllowNone.Value()
		return
	} else if a.HasAll() {
		value = AllowAll.Value()
		return
	}
	value = "("
	for idx, o := range a {
		if idx > 0 {
			value += " "
		}
		value += o.Value()
	}
	value += ")"
	return
}