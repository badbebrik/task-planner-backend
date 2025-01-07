package email

import "fmt"

func GenerateVerificationEmail(code string) string {
	return fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<style>
				body {
					font-family: Arial, sans-serif;
					background-color: #f4f4f9;
					color: #333;
					padding: 20px;
				}
				.container {
					max-width: 600px;
					margin: auto;
					background: #fff;
					padding: 20px;
					border-radius: 10px;
					box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
				}
				h1 {
					color: #4CAF50;
				}
				.code {
					font-size: 24px;
					font-weight: bold;
					color: #333;
					background: #f4f4f9;
					padding: 10px;
					border-radius: 5px;
					display: inline-block;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<h1>Welcome to WhatAmIToDo!</h1>
				<p>Thank you for registering. Your verification code is:</p>
				<p class="code">%s</p>
				<p>Please enter this code to verify your email address.</p>
			</div>
		</body>
		</html>
	`, code)
}
