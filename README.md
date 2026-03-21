# gophermart-loyalty

[![gophermart](https://github.com/YNGXZRIST/gophermart-loyalty/actions/workflows/gophermart.yml/badge.svg)](https://github.com/YNGXZRIST/gophermart-loyalty/actions/workflows/gophermart.yml)
[![go vet test](https://github.com/YNGXZRIST/gophermart-loyalty/actions/workflows/statictest.yml/badge.svg)](https://github.com/YNGXZRIST/gophermart-loyalty/actions/workflows/statictest.yml)
[![coverage](https://raw.githubusercontent.com/YNGXZRIST/gophermart-loyalty/main/.badges/main/coverage.svg)](https://github.com/YNGXZRIST/gophermart-loyalty/actions/workflows/coverage.yml)

Сервис лояльности Gophermart на Go.

## О сервисе

`gophermart-loyalty` — backend для программы лояльности интернет-магазина.  
Сервис:

- регистрирует и авторизует пользователей;
- принимает номера заказов пользователя;
- периодически запрашивает внешний accrual-сервис по статусу начислений;
- начисляет баллы на баланс пользователя;
- обрабатывает списания и хранит историю выводов.

Спецификация HTTP API: [`openapi.yaml`](openapi.yaml).