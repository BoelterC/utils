package path

import (
	"os"
	"path/filepath"
)

// exists returns whether the given file or directory exists
func Exist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//CWD 当前目录
func Cwd() string {
	// 兼容 travis 集成测试
	if os.Getenv("TRAVIS_BUILD_DIR") != "" {
		return os.Getenv("TRAVIS_BUILD_DIR")
	}
	path, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(path)
}

//EnsureDir 保证目录存在 不存在则创建目录
func EnsureDir(dirpath string) (err error) {
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		err = os.MkdirAll(dirpath, 0755)
		if err != nil {
			return err
		}
	}
	return
}
