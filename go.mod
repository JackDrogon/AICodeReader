module github.com/JackDrogon/aicodereader

go 1.24.0

require github.com/sashabaranov/go-openai v1.38.0

require github.com/sabhiram/go-gitignore v0.0.0-20210923224102-525f6e181f06 // indirect

// for deepseek reason, we need to use the following: https://github.com/goodenough227/go-openai/tree/master
// require github.com/goodenough227/go-openai v0.0.0-20240328084325-098180978800
replace github.com/sashabaranov/go-openai => github.com/goodenough227/go-openai v0.0.0-20250313060841-319a8ea883f9
