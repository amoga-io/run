package internal

func GetScriptAndExecute(command, packageName string) error {
	script, err := GetScriptPath(command, packageName)
	if err != nil {
		return err
	}

	if err := ExecuteScript(script); err != nil {
		return err
	}
	return nil
}
