package iqlog

func (vs *varStack) Msg(msg string) {
	// if vs.varCount > maxVars {
	// 	return
	// }
	if vs.level == LevelUnknown {
		return
	}
	if vs.jsonMode {
		vs.writeFinalJSON(msg)
	} else {
		vs.writeFinalConsole(msg)
	}
}
func (vs *varStack) Msgs(msg string, args ...interface{}) {
	// if vs.varCount > maxVars {
	// 	return
	// }
	if vs.level == LevelUnknown {
		return
	}
	if vs.jsonMode {
		vs.writeFinalJSON(msg, args...)
	} else {
		vs.writeFinalConsole(msg, args...)
	}
}
func (vs *varStack) Msgf(format string, args ...interface{}) {
	// if vs.varCount > maxVars {
	// 	return
	// }
	if vs.level == LevelUnknown {
		return
	}
	if len(args) > 0 {
		if vs.jsonMode {
			vs.writeFinalJSONF(format, args...)
		} else {
			vs.writeFinalConsoleF(format, args...)
		}
	} else {
		if vs.jsonMode {
			vs.writeFinalJSON(format)
		} else {
			vs.writeFinalConsole(format)
		}
	}
}
