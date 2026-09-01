<?php declare(strict_types=1); if(requestMethod()!=='POST'){http_response_code(405);exit('Method not allowed.');} Csrf::requireValid($_POST['_csrf']??null);Auth::logout(App::api());redirect('login');
