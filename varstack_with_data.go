package iqlog

func (vs *varStack) Int(name string, i int) *varStack {
	if vs.varCount >= maxVars {
		return vs
	}
	vs.varName[vs.varCount] = name
	vs.varValue[vs.varCount] = i
	vs.varCount++
	return vs
}

func (vs *varStack) Int64(name string, i int64) *varStack {
	if vs.varCount >= maxVars {
		return vs
	}
	vs.varName[vs.varCount] = name
	vs.varValue[vs.varCount] = i
	vs.varCount++
	return vs
}

func (vs *varStack) Str(name string, s string) *varStack {
	if vs.varCount >= maxVars {
		return vs
	}
	vs.varName[vs.varCount] = name
	vs.varValue[vs.varCount] = s
	vs.varCount++
	return vs
}

func (vs *varStack) Float32(name string, f float32) *varStack {
	if vs.varCount >= maxVars {
		return vs
	}
	vs.varName[vs.varCount] = name
	vs.varValue[vs.varCount] = f
	vs.varCount++
	return vs
}

func (vs *varStack) Float64(name string, f float64) *varStack {
	if vs.varCount >= maxVars {
		return vs
	}
	vs.varName[vs.varCount] = name
	vs.varValue[vs.varCount] = f
	vs.varCount++
	return vs
}

func (vs *varStack) Bool(name string, b bool) *varStack {
	if vs.varCount >= maxVars {
		return vs
	}
	vs.varName[vs.varCount] = name
	vs.varValue[vs.varCount] = b
	vs.varCount++
	return vs
}

func (vs *varStack) Any(name string, val any) *varStack {
	if vs.varCount >= maxVars {
		return vs
	}
	vs.varName[vs.varCount] = name
	vs.varValue[vs.varCount] = val
	vs.varCount++
	return vs
}
