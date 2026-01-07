package empty

// IsEmpty 判断字符串是否为空
// 参数:
//   - str: 待判断的字符串
//
// 返回:
//   - true: 字符串为空（长度为0）
//   - false: 字符串不为空
func IsEmpty(str string) bool {
	return len(str) == 0
}

// IsNotEmpty 判断字符串是否不为空
// 参数:
//   - str: 待判断的字符串
//
// 返回:
//   - true: 字符串不为空（长度大于0）
//   - false: 字符串为空
func IsNotEmpty(str string) bool {
	return len(str) > 0
}

// IsEmptyPtr 判断字符串指针是否为空
// 参数:
//   - strPtr: 待判断的字符串指针
//
// 返回:
//   - true: 指针为 nil 或指向的字符串为空（长度为0）
//   - false: 指针不为 nil 且指向的字符串不为空
func IsEmptyPtr(strPtr *string) bool {
	return strPtr == nil || len(*strPtr) == 0
}

// IsNotEmptyPtr 判断字符串指针是否不为空
// 参数:
//   - strPtr: 待判断的字符串指针
//
// 返回:
//   - true: 指针不为 nil 且指向的字符串不为空（长度大于0）
//   - false: 指针为 nil 或指向的字符串为空
func IsNotEmptyPtr(strPtr *string) bool {
	return strPtr != nil && len(*strPtr) > 0
}
