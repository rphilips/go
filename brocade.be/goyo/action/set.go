package action

import (
	qutil "brocade.be/goyo/lib/util"
	qyottadb "brocade.be/goyo/lib/yottadb"
	qprompt "github.com/manifoldco/promptui"
)

func Set(text string) {
	gloref := text
	value, err := qyottadb.G(gloref)
	if err == nil {
		gloref += "=" + value
	}
	answer := gloref
	validate := func(input string) error {
		return nil
	}

	for {
		prompt := qprompt.Prompt{
			Label:     "set",
			Validate:  validate,
			Default:   answer,
			AllowEdit: true,
			Pointer:   qprompt.PipeCursor,
		}
		answer, err := prompt.Run()

		if err != nil {
			answer = ""
		}
		if answer == "" {
			return
		}
		gloref, value, err = qyottadb.SetArg(answer)
		if err != nil {
			qutil.Error(err)
			continue
		}

		err = qyottadb.Set(gloref, value)
		if err != nil {
			qutil.Error(err)
			continue
		}
		break
	}

}
