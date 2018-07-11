# project

Command project is the tp-micro service project.
<br>The framework reference: https://github.com/xiaoenai/tp-micro

## API Desc

### Stat

Stat handler

- URI:
	```
	/project/stat?ts={ts}
	```
- REQUEST:
	```json
	{}
	```
- RESULT:


### Home

Home handler

- URI:
	```
	/project/home
	```
- REQUEST:
- RESULT:
	```json
	{
		"content": ""	// text
	}
	```


### Math_Divide

Divide handler

- URI:
	```
	/project/math/divide
	```
- REQUEST:
	```json
	{
		"a": -0.000000,	// dividend
		"b": -0.000000	// divisor
	}
	```
- RESULT:
	```json
	{
		"c": -0.000000	// quotient
	}
	```




<br>

*This is a project created by `micro gen` command.*

*[About Micro Command](https://github.com/xiaoenai/tp-micro/tree/master/cmd/micro)*
