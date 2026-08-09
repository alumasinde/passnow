<?php
declare(strict_types=1);
final class ApiException extends RuntimeException { public function __construct(string $message,private readonly int $status=0,private readonly array $payload=[]){parent::__construct($message,$status);} public function status():int{return $this->status;} public function payload():array{return $this->payload;} }
