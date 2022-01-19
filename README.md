# load-test

#### Arquivo exemplo para teste:
```
{
    "loops": 1,
    "parallel": 1,
    "variables": [
	{
		"key": "subdomain",
		"value": "page"
	},
    ],
    "cycle": [
        {
		"url": "https://{%VAR:subdomain:ENDVAR%}.example.com/"
        },
        {
		"if": "== {%RESP[0]:STATUS_CODE:ENDRESP%} 200",
		"url": "https://{%VAR:subdomain:ENDVAR%}.example.com/new-account",
		"method": "POST",
		"body_json": {
			"username": "admin",
			"passworld": "admin"
		}
        },
        {
		"if": "== {%PATH[1]:message.details:ENDPATH%} successfully created",
		"url": "https://{%VAR:subdomain:ENDVAR%}.example.com/login",
		"content-type": "application/json",
		"timeout": 30,
		"body_load_file": "/home/user/login-data.json"
        },
    ]
}
```

#### Loops
- Número de vezes que o teste será executado.
- Valor padrão: 1

#### Parallel
- Número de execuções paralelas do ciclo.
- Valor padrão: 1

#### Tipos de variáveis:
- Obter um valor definido no objeto **variables**:
	- {%VAR::ENDVAR%}
- Obter o valor de uma **variável de ambiente**:
	- {%ENV::ENDENV%}
- Obter o valor de um campo em um **response body (JSON)**:
	- {%PATH[*cycle index*]::ENDPATH%}
- Obter o valor de uma **response**:
	- {%RESP[cycle index]::ENDRESP%}
	- Opções:
		- STATUS_CODE

#### Onde usar uma variável:
Se você deseja inserir uma variável com um valor boleano ou inteiro prefira usar os campos: **body** (string JSON) ou **body_load_file**. Usar um campo no objeto **body_json** com uma variável sem as "aspas" vai tornar o arquivo de teste inválido.

- if
- url
- body
- body_json: define o content-type como "application/json"
- body_load_file

#### IF (não é permitido usar no ciclo de índice *ZERO*)
- Estrutura: [*operador*] [*variável*] [*variável/valor fixo*]
- Operadores de igualdade: == (igual), != (diferente)

#### Method
- Valor padrão: **GET**

#### Timeout
- Tempo em segundos.
