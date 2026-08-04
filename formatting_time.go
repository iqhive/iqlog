package iqlog

import (
	"bytes"
	"time"
)

func AddTimeConsoleToBuffer(timeNow time.Time, buffer *bytes.Buffer) {
	year, month, day := timeNow.Date()
	hour, min, sec := timeNow.Clock()
	usec := timeNow.Nanosecond() / 1000

	var consolePrefixFull = [29]byte{
		'[', '0', '0', '0', '0', '-', '0', '0', '-', '0', '0',
		'T', '0', '0', ':', '0', '0', ':', '0', '0', '.',
		'0', '0', '0', '0', '0', '0', ']', ' ',
	}

	// setIntBytes(consolePrefixFull[1:], int64(year), 4)
	// setIntBytes(consolePrefixFull[6:], int64(month), 2)
	// setIntBytes(consolePrefixFull[9:], int64(day), 2)
	// setIntBytes(consolePrefixFull[12:], int64(hour), 2)
	// setIntBytes(consolePrefixFull[15:], int64(min), 2)
	// setIntBytes(consolePrefixFull[18:], int64(sec), 2)
	// setIntBytes(consolePrefixFull[21:], int64(usec), 6)

	val4 := int64(year)
	val3 := (val4 / 10)
	val2 := (val3 / 10)
	val1 := (val2 / 10) % 10
	val2 = val2 % 10
	val3 = val3 % 10
	val4 = val4 % 10
	if (val1) != 0 {
		consolePrefixFull[1] = byte('0' + (val1))
	}
	if (val2) != 0 {
		consolePrefixFull[2] = byte('0' + (val2))
	}
	if (val3) != 0 {
		consolePrefixFull[3] = byte('0' + (val3))
	}
	if (val4) != 0 {
		consolePrefixFull[4] = byte('0' + (val4))
	}

	// setIntBytes(consolePrefixFull[6:], int64(month), 2)
	val2 = int64(month)
	val1 = (val2 / 10) % 10
	val2 = val2 % 10
	if (val1) != 0 {
		consolePrefixFull[6] = byte('0' + (val1))
	}
	if (val2) != 0 {
		consolePrefixFull[7] = byte('0' + (val2))
	}

	// setIntBytes(consolePrefixFull[9:], int64(day), 2)
	val2 = int64(day)
	val1 = (val2 / 10) % 10
	val2 = val2 % 10
	if (val1) != 0 {
		consolePrefixFull[9] = byte('0' + (val1))
	}
	if (val2) != 0 {
		consolePrefixFull[10] = byte('0' + (val2))
	}
	// setIntBytes(consolePrefixFull[12:], int64(hour), 2)
	val2 = int64(hour)
	val1 = (val2 / 10) % 10
	val2 = val2 % 10
	if (val1) != 0 {
		consolePrefixFull[12] = byte('0' + (val1))
	}
	if (val2) != 0 {
		consolePrefixFull[13] = byte('0' + (val2))
	}

	// setIntBytes(consolePrefixFull[15:], int64(min), 2)
	val2 = int64(min)
	val1 = (val2 / 10) % 10
	val2 = val2 % 10
	if (val1) != 0 {
		consolePrefixFull[15] = byte('0' + (val1))
	}
	if (val2) != 0 {
		consolePrefixFull[16] = byte('0' + (val2))
	}

	// setIntBytes(consolePrefixFull[18:], int64(sec), 2)
	val2 = int64(sec)
	val1 = (val2 / 10) % 10
	val2 = val2 % 10
	if (val1) != 0 {
		consolePrefixFull[18] = byte('0' + (val1))
	}
	if (val2) != 0 {
		consolePrefixFull[19] = byte('0' + (val2))
	}

	// setIntBytes(consolePrefixFull[21:], int64(usec), 6)
	val6 := int64(usec)
	val5 := (val6 / 10)
	val4 = (val5 / 10)
	val3 = (val4 / 10)
	val2 = (val3 / 10)
	val1 = (val2 / 10) % 10
	val2 = val2 % 10
	val3 = val3 % 10
	val4 = val4 % 10
	val5 = val5 % 10
	val6 = val6 % 10
	if (val1) != 0 {
		consolePrefixFull[21] = byte('0' + (val1))
	}
	if (val2) != 0 {
		consolePrefixFull[22] = byte('0' + (val2))
	}
	if (val3) != 0 {
		consolePrefixFull[23] = byte('0' + (val3))
	}
	if (val4) != 0 {
		consolePrefixFull[24] = byte('0' + (val4))
	}
	if (val5) != 0 {
		consolePrefixFull[25] = byte('0' + (val5))
	}
	if (val6) != 0 {
		consolePrefixFull[26] = byte('0' + (val6))
	}

	buffer.Write(consolePrefixFull[:29])

}

func AddTimeJSONToBuffer(timeNow time.Time, buffer *bytes.Buffer) {

	year, month, day := timeNow.Date()
	hour, min, sec := timeNow.Clock()
	usec := timeNow.Nanosecond() / 1000

	var prefixArr = [37]byte{
		'{', '"', 't', 'i', 'm', 'e', '"', ':', '"',
		'0', '0', '0', '0', '-', '0', '0', '-', '0', '0',
		'T', '0', '0', ':', '0', '0', ':', '0', '0', '.',
		'0', '0', '0', '0', '0', '0', '"', ',',
	}

	// setIntBytes(prefixArr[9:], int64(year), 4)
	// setIntBytes(prefixArr[14:], int64(month), 2)
	// setIntBytes(prefixArr[17:], int64(day), 2)
	// setIntBytes(prefixArr[20:], int64(hour), 2)
	// setIntBytes(prefixArr[23:], int64(min), 2)
	// setIntBytes(prefixArr[26:], int64(sec), 2)
	// setIntBytes(prefixArr[29:], int64(usec), 6)

	// setIntBytes(prefixArr[9:], int64(year), 4)
	val4 := int64(year)
	val3 := (val4 / 10)
	val2 := (val3 / 10)
	val1 := (val2 / 10) % 10
	if (val1) != 0 {
		prefixArr[9] = byte('0' + (val1))
	}
	if (val2 % 10) != 0 {
		prefixArr[10] = byte('0' + (val2 % 10))
	}
	if (val3 % 10) != 0 {
		prefixArr[11] = byte('0' + (val3 % 10))
	}
	if (val4 % 10) != 0 {
		prefixArr[12] = byte('0' + (val4 % 10))
	}

	// prefixArr[13] = '-'

	// setIntBytes(prefixArr[14:], int64(month), 2)
	val2 = int64(month)
	val1 = (val2 / 10) % 10
	if (val1) != 0 {
		prefixArr[14] = byte('0' + (val1))
	}
	if (val2 % 10) != 0 {
		prefixArr[15] = byte('0' + (val2 % 10))
	}

	// prefixArr[16] = '-'

	// setIntBytes(prefixArr[17:], int64(day), 2)
	val2 = int64(day)
	val1 = (val2 / 10) % 10
	if (val1) != 0 {
		prefixArr[17] = byte('0' + (val1))
	}
	if (val2 % 10) != 0 {
		prefixArr[18] = byte('0' + (val2 % 10))
	}

	// prefixArr[19] = 'T'

	// setIntBytes(prefixArr[20:], int64(hour), 2)
	val2 = int64(hour)
	val1 = (val2 / 10) % 10
	if (val1) != 0 {
		prefixArr[20] = byte('0' + (val1))
	}
	if (val2 % 10) != 0 {
		prefixArr[21] = byte('0' + (val2 % 10))
	}

	// prefixArr[22] = ':'

	// setIntBytes(prefixArr[23:], int64(min), 2)
	val2 = int64(min)
	val1 = (val2 / 10) % 10
	if (val1) != 0 {
		prefixArr[23] = byte('0' + (val1))
	}
	if (val2 % 10) != 0 {
		prefixArr[24] = byte('0' + (val2 % 10))
	}

	// prefixArr[25] = ':'

	// setIntBytes(prefixArr[26:], int64(sec), 2)
	val2 = int64(sec)
	val1 = (val2 / 10) % 10
	if (val1) != 0 {
		prefixArr[26] = byte('0' + (val1))
	}
	if (val2 % 10) != 0 {
		prefixArr[27] = byte('0' + (val2 % 10))
	}

	// prefixArr[28] = '.'

	// setIntBytes(prefixArr[29:], int64(usec), 6)
	val6 := int64(usec)
	val5 := (val6 / 10)
	val4 = (val5 / 10)
	val3 = (val4 / 10)
	val2 = (val3 / 10)
	val1 = (val2 / 10) % 10
	// if (val1) != 0 {
	// 	prefixArr[29] = byte('0' + (val1))
	// }
	// if (val2 % 10) != 0 {
	// 	prefixArr[30] = byte('0' + (val2 % 10))
	// }
	// if (val3 % 10) != 0 {
	// 	prefixArr[31] = byte('0' + (val3 % 10))
	// }
	// if (val4 % 10) != 0 {
	// 	prefixArr[32] = byte('0' + (val4 % 10))
	// }
	// if (val5 % 10) != 0 {
	// 	prefixArr[33] = byte('0' + (val5 % 10))
	// }
	// if (val6 % 10) != 0 {
	// 	prefixArr[34] = byte('0' + (val6 % 10))
	// }

	if (val1) != 0 {
		prefixArr[29] = byte('0' + (val1))
	}
	val2 = val2 % 10
	if (val2) != 0 {
		prefixArr[30] = byte('0' + (val2))
	}
	val3 = val3 % 10
	if (val3) != 0 {
		prefixArr[31] = byte('0' + (val3))
	}
	val4 = val4 % 10
	if (val4) != 0 {
		prefixArr[32] = byte('0' + (val4))
	}
	val5 = val5 % 10
	if (val5) != 0 {
		prefixArr[33] = byte('0' + (val5))
	}
	val6 = val6 % 10
	if (val6) != 0 {
		prefixArr[34] = byte('0' + (val6))
	}

	buffer.Write(prefixArr[:37])
}

func AddTimeConsoleAppend(timeNow time.Time, output *[]byte) int {
	year, month, day := timeNow.Date()
	hour, min, sec := timeNow.Clock()
	usec := timeNow.Nanosecond() / 1000

	var consolePrefixFull = [29]byte{
		'[', '0', '0', '0', '0', '-', '0', '0', '-', '0', '0',
		'T', '0', '0', ':', '0', '0', ':', '0', '0', '.',
		'0', '0', '0', '0', '0', '0', ']', ' ',
	}

	// setIntBytes(consolePrefixFull[1:], int64(year), 4)
	// setIntBytes(consolePrefixFull[6:], int64(month), 2)
	// setIntBytes(consolePrefixFull[9:], int64(day), 2)
	// setIntBytes(consolePrefixFull[12:], int64(hour), 2)
	// setIntBytes(consolePrefixFull[15:], int64(min), 2)
	// setIntBytes(consolePrefixFull[18:], int64(sec), 2)
	// setIntBytes(consolePrefixFull[21:], int64(usec), 6)

	val4 := int64(year)
	val3 := (val4 / 10)
	val2 := (val3 / 10)
	val1 := (val2 / 10) % 10
	val2 = val2 % 10
	val3 = val3 % 10
	val4 = val4 % 10
	if (val1) != 0 {
		consolePrefixFull[1] = byte('0' + (val1))
	}
	if (val2) != 0 {
		consolePrefixFull[2] = byte('0' + (val2))
	}
	if (val3) != 0 {
		consolePrefixFull[3] = byte('0' + (val3))
	}
	if (val4) != 0 {
		consolePrefixFull[4] = byte('0' + (val4))
	}

	// setIntBytes(consolePrefixFull[6:], int64(month), 2)
	val2 = int64(month)
	val1 = (val2 / 10) % 10
	val2 = val2 % 10
	if (val1) != 0 {
		consolePrefixFull[6] = byte('0' + (val1))
	}
	if (val2) != 0 {
		consolePrefixFull[7] = byte('0' + (val2))
	}

	// setIntBytes(consolePrefixFull[9:], int64(day), 2)
	val2 = int64(day)
	val1 = (val2 / 10) % 10
	val2 = val2 % 10
	if (val1) != 0 {
		consolePrefixFull[9] = byte('0' + (val1))
	}
	if (val2) != 0 {
		consolePrefixFull[10] = byte('0' + (val2))
	}
	// setIntBytes(consolePrefixFull[12:], int64(hour), 2)
	val2 = int64(hour)
	val1 = (val2 / 10) % 10
	val2 = val2 % 10
	if (val1) != 0 {
		consolePrefixFull[12] = byte('0' + (val1))
	}
	if (val2) != 0 {
		consolePrefixFull[13] = byte('0' + (val2))
	}

	// setIntBytes(consolePrefixFull[15:], int64(min), 2)
	val2 = int64(min)
	val1 = (val2 / 10) % 10
	val2 = val2 % 10
	if (val1) != 0 {
		consolePrefixFull[15] = byte('0' + (val1))
	}
	if (val2) != 0 {
		consolePrefixFull[16] = byte('0' + (val2))
	}

	// setIntBytes(consolePrefixFull[18:], int64(sec), 2)
	val2 = int64(sec)
	val1 = (val2 / 10) % 10
	val2 = val2 % 10
	if (val1) != 0 {
		consolePrefixFull[18] = byte('0' + (val1))
	}
	if (val2) != 0 {
		consolePrefixFull[19] = byte('0' + (val2))
	}

	// setIntBytes(consolePrefixFull[21:], int64(usec), 6)
	val6 := int64(usec)
	val5 := (val6 / 10)
	val4 = (val5 / 10)
	val3 = (val4 / 10)
	val2 = (val3 / 10)
	val1 = (val2 / 10) % 10
	val2 = val2 % 10
	val3 = val3 % 10
	val4 = val4 % 10
	val5 = val5 % 10
	val6 = val6 % 10
	if (val1) != 0 {
		consolePrefixFull[21] = byte('0' + (val1))
	}
	if (val2) != 0 {
		consolePrefixFull[22] = byte('0' + (val2))
	}
	if (val3) != 0 {
		consolePrefixFull[23] = byte('0' + (val3))
	}
	if (val4) != 0 {
		consolePrefixFull[24] = byte('0' + (val4))
	}
	if (val5) != 0 {
		consolePrefixFull[25] = byte('0' + (val5))
	}
	if (val6) != 0 {
		consolePrefixFull[26] = byte('0' + (val6))
	}

	*output = append(*output, consolePrefixFull[:29]...)
	return 29
}

func AddTimeJSONAppend(timeNow time.Time, output *[]byte) int {
	year, month, day := timeNow.Date()
	hour, min, sec := timeNow.Clock()
	usec := timeNow.Nanosecond() / 1000

	var prefixArr = [37]byte{
		'{', '"', 't', 'i', 'm', 'e', '"', ':', '"',
		'0', '0', '0', '0', '-', '0', '0', '-', '0', '0',
		'T', '0', '0', ':', '0', '0', ':', '0', '0', '.',
		'0', '0', '0', '0', '0', '0', '"', ',',
	}

	// setIntBytes(prefixArr[9:], int64(year), 4)
	// setIntBytes(prefixArr[14:], int64(month), 2)
	// setIntBytes(prefixArr[17:], int64(day), 2)
	// setIntBytes(prefixArr[20:], int64(hour), 2)
	// setIntBytes(prefixArr[23:], int64(min), 2)
	// setIntBytes(prefixArr[26:], int64(sec), 2)
	// setIntBytes(prefixArr[29:], int64(usec), 6)

	// setIntBytes(prefixArr[9:], int64(year), 4)
	val4 := int64(year)
	val3 := (val4 / 10)
	val2 := (val3 / 10)
	val1 := (val2 / 10) % 10
	if (val1) != 0 {
		prefixArr[9] = byte('0' + (val1))
	}
	if (val2 % 10) != 0 {
		prefixArr[10] = byte('0' + (val2 % 10))
	}
	if (val3 % 10) != 0 {
		prefixArr[11] = byte('0' + (val3 % 10))
	}
	if (val4 % 10) != 0 {
		prefixArr[12] = byte('0' + (val4 % 10))
	}

	// prefixArr[13] = '-'

	// setIntBytes(prefixArr[14:], int64(month), 2)
	val2 = int64(month)
	val1 = (val2 / 10) % 10
	if (val1) != 0 {
		prefixArr[14] = byte('0' + (val1))
	}
	if (val2 % 10) != 0 {
		prefixArr[15] = byte('0' + (val2 % 10))
	}

	// prefixArr[16] = '-'

	// setIntBytes(prefixArr[17:], int64(day), 2)
	val2 = int64(day)
	val1 = (val2 / 10) % 10
	if (val1) != 0 {
		prefixArr[17] = byte('0' + (val1))
	}
	if (val2 % 10) != 0 {
		prefixArr[18] = byte('0' + (val2 % 10))
	}

	// prefixArr[19] = 'T'

	// setIntBytes(prefixArr[20:], int64(hour), 2)
	val2 = int64(hour)
	val1 = (val2 / 10) % 10
	if (val1) != 0 {
		prefixArr[20] = byte('0' + (val1))
	}
	if (val2 % 10) != 0 {
		prefixArr[21] = byte('0' + (val2 % 10))
	}

	// prefixArr[22] = ':'

	// setIntBytes(prefixArr[23:], int64(min), 2)
	val2 = int64(min)
	val1 = (val2 / 10) % 10
	if (val1) != 0 {
		prefixArr[23] = byte('0' + (val1))
	}
	if (val2 % 10) != 0 {
		prefixArr[24] = byte('0' + (val2 % 10))
	}

	// prefixArr[25] = ':'

	// setIntBytes(prefixArr[26:], int64(sec), 2)
	val2 = int64(sec)
	val1 = (val2 / 10) % 10
	if (val1) != 0 {
		prefixArr[26] = byte('0' + (val1))
	}
	if (val2 % 10) != 0 {
		prefixArr[27] = byte('0' + (val2 % 10))
	}

	// prefixArr[28] = '.'

	// setIntBytes(prefixArr[29:], int64(usec), 6)
	val6 := int64(usec)
	val5 := (val6 / 10)
	val4 = (val5 / 10)
	val3 = (val4 / 10)
	val2 = (val3 / 10)
	val1 = (val2 / 10) % 10
	// if (val1) != 0 {
	// 	prefixArr[29] = byte('0' + (val1))
	// }
	// if (val2 % 10) != 0 {
	// 	prefixArr[30] = byte('0' + (val2 % 10))
	// }
	// if (val3 % 10) != 0 {
	// 	prefixArr[31] = byte('0' + (val3 % 10))
	// }
	// if (val4 % 10) != 0 {
	// 	prefixArr[32] = byte('0' + (val4 % 10))
	// }
	// if (val5 % 10) != 0 {
	// 	prefixArr[33] = byte('0' + (val5 % 10))
	// }
	// if (val6 % 10) != 0 {
	// 	prefixArr[34] = byte('0' + (val6 % 10))
	// }

	if (val1) != 0 {
		prefixArr[29] = byte('0' + (val1))
	}
	val2 = val2 % 10
	if (val2) != 0 {
		prefixArr[30] = byte('0' + (val2))
	}
	val3 = val3 % 10
	if (val3) != 0 {
		prefixArr[31] = byte('0' + (val3))
	}
	val4 = val4 % 10
	if (val4) != 0 {
		prefixArr[32] = byte('0' + (val4))
	}
	val5 = val5 % 10
	if (val5) != 0 {
		prefixArr[33] = byte('0' + (val5))
	}
	val6 = val6 % 10
	if (val6) != 0 {
		prefixArr[34] = byte('0' + (val6))
	}

	// Use the comma
	*output = append(*output, prefixArr[:37]...)
	return 37
}

func AddTimeConsoleInPlaceCopy(timeNow time.Time, output []byte) int {

	year, month, day := timeNow.Date()
	hour, min, sec := timeNow.Clock()
	usec := timeNow.Nanosecond() / 1000

	var consolePrefixFull = [29]byte{
		'[', '0', '0', '0', '0', '-', '0', '0', '-', '0', '0',
		'T', '0', '0', ':', '0', '0', ':', '0', '0', '.',
		'0', '0', '0', '0', '0', '0', ']', ' ',
	}

	// setIntBytes(consolePrefixFull[1:], int64(year), 4)
	// setIntBytes(consolePrefixFull[6:], int64(month), 2)
	// setIntBytes(consolePrefixFull[9:], int64(day), 2)
	// setIntBytes(consolePrefixFull[12:], int64(hour), 2)
	// setIntBytes(consolePrefixFull[15:], int64(min), 2)
	// setIntBytes(consolePrefixFull[18:], int64(sec), 2)
	// setIntBytes(consolePrefixFull[21:], int64(usec), 6)
	// setIntBytes(consolePrefixFull[1:], int64(year), 4)

	val4 := int64(year)
	val3 := (val4 / 10)
	val2 := (val3 / 10)
	val1 := (val2 / 10) % 10
	val2 = val2 % 10
	val3 = val3 % 10
	val4 = val4 % 10
	if (val1) != 0 {
		consolePrefixFull[1] = byte('0' + (val1))
	}
	if (val2) != 0 {
		consolePrefixFull[2] = byte('0' + (val2))
	}
	if (val3) != 0 {
		consolePrefixFull[3] = byte('0' + (val3))
	}
	if (val4) != 0 {
		consolePrefixFull[4] = byte('0' + (val4))
	}

	// setIntBytes(consolePrefixFull[6:], int64(month), 2)
	val2 = int64(month)
	val1 = (val2 / 10) % 10
	val2 = val2 % 10
	if (val1) != 0 {
		consolePrefixFull[6] = byte('0' + (val1))
	}
	if (val2) != 0 {
		consolePrefixFull[7] = byte('0' + (val2))
	}

	// setIntBytes(consolePrefixFull[9:], int64(day), 2)
	val2 = int64(day)
	val1 = (val2 / 10) % 10
	val2 = val2 % 10
	if (val1) != 0 {
		consolePrefixFull[9] = byte('0' + (val1))
	}
	if (val2) != 0 {
		consolePrefixFull[10] = byte('0' + (val2))
	}
	// setIntBytes(consolePrefixFull[12:], int64(hour), 2)
	val2 = int64(hour)
	val1 = (val2 / 10) % 10
	val2 = val2 % 10
	if (val1) != 0 {
		consolePrefixFull[12] = byte('0' + (val1))
	}
	if (val2) != 0 {
		consolePrefixFull[13] = byte('0' + (val2))
	}

	// setIntBytes(consolePrefixFull[15:], int64(min), 2)
	val2 = int64(min)
	val1 = (val2 / 10) % 10
	val2 = val2 % 10
	if (val1) != 0 {
		consolePrefixFull[15] = byte('0' + (val1))
	}
	if (val2) != 0 {
		consolePrefixFull[16] = byte('0' + (val2))
	}

	// setIntBytes(consolePrefixFull[18:], int64(sec), 2)
	val2 = int64(sec)
	val1 = (val2 / 10) % 10
	val2 = val2 % 10
	if (val1) != 0 {
		consolePrefixFull[18] = byte('0' + (val1))
	}
	if (val2) != 0 {
		consolePrefixFull[19] = byte('0' + (val2))
	}

	// setIntBytes(consolePrefixFull[21:], int64(usec), 6)
	val6 := int64(usec)
	val5 := (val6 / 10)
	val4 = (val5 / 10)
	val3 = (val4 / 10)
	val2 = (val3 / 10)
	val1 = (val2 / 10) % 10
	val2 = val2 % 10
	val3 = val3 % 10
	val4 = val4 % 10
	val5 = val5 % 10
	val6 = val6 % 10
	if (val1) != 0 {
		consolePrefixFull[21] = byte('0' + (val1))
	}
	if (val2) != 0 {
		consolePrefixFull[22] = byte('0' + (val2))
	}
	if (val3) != 0 {
		consolePrefixFull[23] = byte('0' + (val3))
	}
	if (val4) != 0 {
		consolePrefixFull[24] = byte('0' + (val4))
	}
	if (val5) != 0 {
		consolePrefixFull[25] = byte('0' + (val5))
	}
	if (val6) != 0 {
		consolePrefixFull[26] = byte('0' + (val6))
	}

	copy(output[0:], consolePrefixFull[:29])
	return 29
}

//
// commented out because it is slightly slower (+5% ns) than AddTimeJSONInPlaceCopy
//
// func AddTimeConsoleInPlaceOverwrite(timeNow time.Time, output []byte) int {

// 	year, month, day := timeNow.Date()
// 	hour, min, sec := timeNow.Clock()
// 	usec := timeNow.Nanosecond() / 1000

// 	var consolePrefixFull = [29]byte{
// 		'[', '0', '0', '0', '0', '-', '0', '0', '-', '0', '0',
// 		'T', '0', '0', ':', '0', '0', ':', '0', '0', '.',
// 		'0', '0', '0', '0', '0', '0', ']', ' ',
// 	}

// 	copy(output[0:], consolePrefixFull[:29])

// 	// setIntBytes(consolePrefixFull[1:], int64(year), 4)
// 	val4 := int64(year)
// 	val3 := (val4 / 10)
// 	val2 := (val3 / 10)
// 	val1 := (val2 / 10) % 10
// 	val2 = val2 % 10
// 	val3 = val3 % 10
// 	val4 = val4 % 10
// 	if (val1) != 0 {
// 		output[1] = byte('0' + (val1))
// 	}
// 	if (val2) != 0 {
// 		output[2] = byte('0' + (val2))
// 	}
// 	if (val3) != 0 {
// 		output[3] = byte('0' + (val3))
// 	}
// 	if (val4) != 0 {
// 		output[4] = byte('0' + (val4))
// 	}

// 	// setIntBytes(consolePrefixFull[6:], int64(month), 2)
// 	val2 = int64(month)
// 	val1 = (val2 / 10) % 10
// 	val2 = val2 % 10
// 	if (val1) != 0 {
// 		output[6] = byte('0' + (val1))
// 	}
// 	if (val2) != 0 {
// 		output[7] = byte('0' + (val2))
// 	}

// 	// setIntBytes(consolePrefixFull[9:], int64(day), 2)
// 	val2 = int64(day)
// 	val1 = (val2 / 10) % 10
// 	val2 = val2 % 10
// 	if (val1) != 0 {
// 		output[9] = byte('0' + (val1))
// 	}
// 	if (val2) != 0 {
// 		output[10] = byte('0' + (val2))
// 	}
// 	// setIntBytes(consolePrefixFull[12:], int64(hour), 2)
// 	val2 = int64(hour)
// 	val1 = (val2 / 10) % 10
// 	val2 = val2 % 10
// 	if (val1) != 0 {
// 		output[12] = byte('0' + (val1))
// 	}
// 	if (val2) != 0 {
// 		output[13] = byte('0' + (val2))
// 	}

// 	// setIntBytes(consolePrefixFull[15:], int64(min), 2)
// 	val2 = int64(min)
// 	val1 = (val2 / 10) % 10
// 	val2 = val2 % 10
// 	if (val1) != 0 {
// 		output[15] = byte('0' + (val1))
// 	}
// 	if (val2) != 0 {
// 		output[16] = byte('0' + (val2))
// 	}

// 	// setIntBytes(consolePrefixFull[18:], int64(sec), 2)
// 	val2 = int64(sec)
// 	val1 = (val2 / 10) % 10
// 	val2 = val2 % 10
// 	if (val1) != 0 {
// 		output[18] = byte('0' + (val1))
// 	}
// 	if (val2) != 0 {
// 		output[19] = byte('0' + (val2))
// 	}

// 	// setIntBytes(consolePrefixFull[21:], int64(usec), 6)
// 	val6 := int64(usec)
// 	val5 := (val6 / 10)
// 	val4 = (val5 / 10)
// 	val3 = (val4 / 10)
// 	val2 = (val3 / 10)
// 	val1 = (val2 / 10) % 10
// 	val2 = val2 % 10
// 	val3 = val3 % 10
// 	val4 = val4 % 10
// 	val5 = val5 % 10
// 	val6 = val6 % 10
// 	if (val1) != 0 {
// 		output[21] = byte('0' + (val1))
// 	}
// 	if (val2) != 0 {
// 		output[22] = byte('0' + (val2))
// 	}
// 	if (val3) != 0 {
// 		output[23] = byte('0' + (val3))
// 	}
// 	if (val4) != 0 {
// 		output[24] = byte('0' + (val4))
// 	}
// 	if (val5) != 0 {
// 		output[25] = byte('0' + (val5))
// 	}
// 	if (val6) != 0 {
// 		output[26] = byte('0' + (val6))
// 	}

// 	// copy(output[0:], consolePrefixFull[:29])
// 	return 29
// }

// func AddTimeJSONInPlaceCopySlow(timeNow time.Time, output []byte) int {

// 	year, month, day := timeNow.Date()
// 	hour, min, sec := timeNow.Clock()
// 	usec := timeNow.Nanosecond() / 1000

// 	var prefixArr = [37]byte{
// 		'{', '"', 't', 'i', 'm', 'e', '"', ':', '"',
// 		'0', '0', '0', '0', '-', '0', '0', '-', '0', '0',
// 		'T', '0', '0', ':', '0', '0', ':', '0', '0', '.',
// 		'0', '0', '0', '0', '0', '0', '"', ',',
// 	}

// 	setIntBytes(prefixArr[9:], int64(year), 4)
// 	setIntBytes(prefixArr[14:], int64(month), 2)
// 	setIntBytes(prefixArr[17:], int64(day), 2)
// 	setIntBytes(prefixArr[20:], int64(hour), 2)
// 	setIntBytes(prefixArr[23:], int64(min), 2)
// 	setIntBytes(prefixArr[26:], int64(sec), 2)
// 	setIntBytes(prefixArr[29:], int64(usec), 6)

// 	// Use the comma
// 	copy(output[0:], prefixArr[:37])
// 	return 37

// }

func AddTimeJSONInPlaceCopy(timeNow time.Time, output []byte) int {

	year, month, day := timeNow.Date()
	hour, min, sec := timeNow.Clock()
	usec := timeNow.Nanosecond() / 1000

	var prefixArr = [37]byte{
		'{', '"', 't', 'i', 'm', 'e', '"', ':', '"',
		'0', '0', '0', '0', '-', '0', '0', '-', '0', '0',
		'T', '0', '0', ':', '0', '0', ':', '0', '0', '.',
		'0', '0', '0', '0', '0', '0', '"', ',',
	}

	// setIntBytes(prefixArr[9:], int64(year), 4)
	val4 := int64(year)
	val3 := (val4 / 10)
	val2 := (val3 / 10)
	val1 := (val2 / 10) % 10
	if (val1) != 0 {
		prefixArr[9] = byte('0' + (val1))
	}
	if (val2 % 10) != 0 {
		prefixArr[10] = byte('0' + (val2 % 10))
	}
	if (val3 % 10) != 0 {
		prefixArr[11] = byte('0' + (val3 % 10))
	}
	if (val4 % 10) != 0 {
		prefixArr[12] = byte('0' + (val4 % 10))
	}

	// prefixArr[13] = '-'

	// setIntBytes(prefixArr[14:], int64(month), 2)
	val2 = int64(month)
	val1 = (val2 / 10) % 10
	if (val1) != 0 {
		prefixArr[14] = byte('0' + (val1))
	}
	if (val2 % 10) != 0 {
		prefixArr[15] = byte('0' + (val2 % 10))
	}

	// prefixArr[16] = '-'

	// setIntBytes(prefixArr[17:], int64(day), 2)
	val2 = int64(day)
	val1 = (val2 / 10) % 10
	if (val1) != 0 {
		prefixArr[17] = byte('0' + (val1))
	}
	if (val2 % 10) != 0 {
		prefixArr[18] = byte('0' + (val2 % 10))
	}

	// prefixArr[19] = 'T'

	// setIntBytes(prefixArr[20:], int64(hour), 2)
	val2 = int64(hour)
	val1 = (val2 / 10) % 10
	if (val1) != 0 {
		prefixArr[20] = byte('0' + (val1))
	}
	if (val2 % 10) != 0 {
		prefixArr[21] = byte('0' + (val2 % 10))
	}

	// prefixArr[22] = ':'

	// setIntBytes(prefixArr[23:], int64(min), 2)
	val2 = int64(min)
	val1 = (val2 / 10) % 10
	if (val1) != 0 {
		prefixArr[23] = byte('0' + (val1))
	}
	if (val2 % 10) != 0 {
		prefixArr[24] = byte('0' + (val2 % 10))
	}

	// prefixArr[25] = ':'

	// setIntBytes(prefixArr[26:], int64(sec), 2)
	val2 = int64(sec)
	val1 = (val2 / 10) % 10
	if (val1) != 0 {
		prefixArr[26] = byte('0' + (val1))
	}
	if (val2 % 10) != 0 {
		prefixArr[27] = byte('0' + (val2 % 10))
	}

	// prefixArr[28] = '.'

	// setIntBytes(prefixArr[29:], int64(usec), 6)
	val6 := int64(usec)
	val5 := (val6 / 10)
	val4 = (val5 / 10)
	val3 = (val4 / 10)
	val2 = (val3 / 10)
	val1 = (val2 / 10) % 10
	// if (val1) != 0 {
	// 	prefixArr[29] = byte('0' + (val1))
	// }
	// if (val2 % 10) != 0 {
	// 	prefixArr[30] = byte('0' + (val2 % 10))
	// }
	// if (val3 % 10) != 0 {
	// 	prefixArr[31] = byte('0' + (val3 % 10))
	// }
	// if (val4 % 10) != 0 {
	// 	prefixArr[32] = byte('0' + (val4 % 10))
	// }
	// if (val5 % 10) != 0 {
	// 	prefixArr[33] = byte('0' + (val5 % 10))
	// }
	// if (val6 % 10) != 0 {
	// 	prefixArr[34] = byte('0' + (val6 % 10))
	// }

	if (val1) != 0 {
		prefixArr[29] = byte('0' + (val1))
	}
	val2 = val2 % 10
	if (val2) != 0 {
		prefixArr[30] = byte('0' + (val2))
	}
	val3 = val3 % 10
	if (val3) != 0 {
		prefixArr[31] = byte('0' + (val3))
	}
	val4 = val4 % 10
	if (val4) != 0 {
		prefixArr[32] = byte('0' + (val4))
	}
	val5 = val5 % 10
	if (val5) != 0 {
		prefixArr[33] = byte('0' + (val5))
	}
	val6 = val6 % 10
	if (val6) != 0 {
		prefixArr[34] = byte('0' + (val6))
	}

	// Use the comma
	copy(output[0:], prefixArr[:37])
	return 37

}

// // AddTimeJSONInPlaceOverwrite overwrites the time in the output buffer
// // It is slightly slower (+5% ns) than AddTimeJSONInPlaceCopy
// func AddTimeJSONInPlaceOverwrite(timeNow time.Time, output []byte) int {

// 	year, month, day := timeNow.Date()
// 	hour, min, sec := timeNow.Clock()
// 	usec := timeNow.Nanosecond() / 1000

// 	var prefixArr = [37]byte{
// 		'{', '"', 't', 'i', 'm', 'e', '"', ':', '"',
// 		'0', '0', '0', '0', '-', '0', '0', '-', '0', '0',
// 		'T', '0', '0', ':', '0', '0', ':', '0', '0', '.',
// 		'0', '0', '0', '0', '0', '0', '"', ',',
// 	}

// 	// output[0] = '{'
// 	// output[1] = '"'
// 	// output[2] = 't'
// 	// output[3] = 'i'
// 	// output[4] = 'm'
// 	// output[5] = 'e'
// 	// output[6] = '"'
// 	// output[7] = ':'
// 	// output[8] = '"'
// 	copy(output[0:], prefixArr[0:37])

// 	// setIntBytes(prefixArr[9:], int64(year), 4)
// 	val4 := int64(year)
// 	val3 := (val4 / 10)
// 	val2 := (val3 / 10)
// 	val1 := (val2 / 10) % 10
// 	if (val1) != 0 {
// 		output[9] = byte('0' + (val1))
// 	}
// 	if (val2 % 10) != 0 {
// 		output[10] = byte('0' + (val2 % 10))
// 	}
// 	if (val3 % 10) != 0 {
// 		output[11] = byte('0' + (val3 % 10))
// 	}
// 	if (val4 % 10) != 0 {
// 		output[12] = byte('0' + (val4 % 10))
// 	}

// 	// output[13] = '-'

// 	// setIntBytes(prefixArr[14:], int64(month), 2)
// 	val2 = int64(month)
// 	val1 = (val2 / 10) % 10
// 	if (val1) != 0 {
// 		output[14] = byte('0' + (val1))
// 	}
// 	if (val2 % 10) != 0 {
// 		output[15] = byte('0' + (val2 % 10))
// 	}

// 	// output[16] = '-'

// 	// setIntBytes(prefixArr[17:], int64(day), 2)
// 	val2 = int64(day)
// 	val1 = (val2 / 10) % 10
// 	if (val1) != 0 {
// 		output[17] = byte('0' + (val1))
// 	}
// 	if (val2 % 10) != 0 {
// 		output[18] = byte('0' + (val2 % 10))
// 	}

// 	// output[19] = 'T'

// 	// setIntBytes(prefixArr[20:], int64(hour), 2)
// 	val2 = int64(hour)
// 	val1 = (val2 / 10) % 10
// 	if (val1) != 0 {
// 		output[20] = byte('0' + (val1))
// 	}
// 	if (val2 % 10) != 0 {
// 		output[21] = byte('0' + (val2 % 10))
// 	}

// 	// output[22] = ':'

// 	// setIntBytes(prefixArr[23:], int64(min), 2)
// 	val2 = int64(min)
// 	val1 = (val2 / 10) % 10
// 	if (val1) != 0 {
// 		output[23] = byte('0' + (val1))
// 	}
// 	if (val2 % 10) != 0 {
// 		output[24] = byte('0' + (val2 % 10))
// 	}

// 	// output[25] = ':'

// 	// setIntBytes(prefixArr[26:], int64(sec), 2)
// 	val2 = int64(sec)
// 	val1 = (val2 / 10) % 10
// 	if (val1) != 0 {
// 		output[26] = byte('0' + (val1))
// 	}
// 	if (val2 % 10) != 0 {
// 		output[27] = byte('0' + (val2 % 10))
// 	}

// 	// output[28] = '.'

// 	// setIntBytes(prefixArr[29:], int64(usec), 6)
// 	val6 := int64(usec)
// 	val5 := (val6 / 10)
// 	val4 = (val5 / 10)
// 	val3 = (val4 / 10)
// 	val2 = (val3 / 10)
// 	val1 = (val2 / 10) % 10
// 	// if (val1) != 0 {
// 	// 	output[29] = byte('0' + (val1))
// 	// }
// 	// if (val2 % 10) != 0 {
// 	// 	output[30] = byte('0' + (val2 % 10))
// 	// }
// 	// if (val3 % 10) != 0 {
// 	// 	output[31] = byte('0' + (val3 % 10))
// 	// }
// 	// if (val4 % 10) != 0 {
// 	// 	output[32] = byte('0' + (val4 % 10))
// 	// }
// 	// if (val5 % 10) != 0 {
// 	// 	output[33] = byte('0' + (val5 % 10))
// 	// }
// 	// if (val6 % 10) != 0 {
// 	// 	output[34] = byte('0' + (val6 % 10))
// 	// }

// 	if (val1) != 0 {
// 		output[29] = byte('0' + (val1))
// 	}
// 	val2 = val2 % 10
// 	if (val2) != 0 {
// 		output[30] = byte('0' + (val2))
// 	}
// 	val3 = val3 % 10
// 	if (val3) != 0 {
// 		output[31] = byte('0' + (val3))
// 	}
// 	val4 = val4 % 10
// 	if (val4) != 0 {
// 		output[32] = byte('0' + (val4))
// 	}
// 	val5 = val5 % 10
// 	if (val5) != 0 {
// 		output[33] = byte('0' + (val5))
// 	}
// 	val6 = val6 % 10
// 	if (val6) != 0 {
// 		output[34] = byte('0' + (val6))
// 	}

// 	// output[35] = '"'
// 	// output[36] = ','

// 	return 37

// }
