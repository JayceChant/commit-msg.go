package validator

import (
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/JayceChant/commit-msg/state"

)

const (
	mergePrefix   = "Merge "
	revertPattern = `^(Revert|revert)(:| ).+`
	headerPattern = `^((fixup! |squash! )?(\w+)(?:\(([^\)\s]+)\))?: (.+))(?:\n|$)`
)

// Validate ...
func Validate(file string) {
	msg := getMsg(file)

	defer func() {
		err := recover()
		state, ok := err.(state.State)
		if !ok {
			panic(err)
		}

		if state.IsNormal() {
			os.Exit(0)
		} else {
			os.Exit(int(state))
		}
	}()
	validateMsg(msg)
}

func getMsg(path string) string {
	if path == "" {
		state.ArgumentMissing.LogAndExit()
	}

	f, err := os.Stat(path)
	if err != nil {
		log.Println(err)
		state.FileMissing.LogAndExit(path)
	}

	if f.IsDir() {
		log.Println(path, "is not a file.")
		state.FileMissing.LogAndExit(path)
	}

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)
		state.ReadError.LogAndExit(path)
	}

	return string(buf)
}

func checkEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}

func checkType(typ string) {
	for _, t := range TypeList {
		if typ == t {
			return
		}
	}
	state.WrongType.LogAndExit(typ, Types)
}

func checkHeader(header string) {
	if checkEmpty(header) {
		state.EmptyHeader.LogAndExit()
	}

	re := regexp.MustCompile(headerPattern)
	groups := re.FindStringSubmatch(header)

	if groups == nil || checkEmpty(groups[5]) {
		state.BadHeaderFormat.LogAndExit(header)
	}

	typ := groups[3]
	checkType(typ)

	isFixupOrSquash := (groups[2] != "")
	// TODO: 根据配置对scope检查
	// scope := groups[4]
	// TODO: 根据规则对subject检查
	// subject := groups[5]

	length := len(header)
	if length > Config.LineLimit &&
		!(isFixupOrSquash || typ == "revert" || typ == "Revert") {
		state.LineOverLong.LogAndExit(length, Config.LineLimit, header)
	}
}

func checkBody(body string) {
	if checkEmpty(body) {
		if Config.BodyRequired {
			state.BodyMissing.LogAndExit()
		} else {
			state.Validated.LogAndExit()
		}
	}

	if !checkEmpty(strings.SplitN(body, "\n", 2)[0]) {
		state.NoBlankLineBeforeBody.LogAndExit()
	}

	for _, line := range strings.Split(body, "\n") {
		length := len(line)
		if length > Config.LineLimit {
			state.LineOverLong.LogAndExit(length, Config.LineLimit, line)
		}
	}
}

func validateMsg(msg string) {
	if checkEmpty(msg) {
		state.EmptyMessage.LogAndExit()
	}

	if strings.HasPrefix(msg, mergePrefix) {
		state.Merge.LogAndExit()
	}

	sections := strings.SplitN(msg, "\n", 2)

	if m, _ := regexp.MatchString(revertPattern, sections[0]); !m {
		checkHeader(sections[0])
	}

	if len(sections) == 2 {
		checkBody(sections[1])
	} else if Config.BodyRequired {
		state.BodyMissing.LogAndExit()
	}

	state.Validated.LogAndExit()
}