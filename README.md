# gophermart-loyalty

[![gophermart](https://img.shields.io/github/actions/workflow/status/YNGXZRIST/gophermart-loyalty/gophermart.yml?branch=main&label=gophermart&logo=github)](https://github.com/YNGXZRIST/gophermart-loyalty/actions/workflows/gophermart.yml?query=branch%3Amain)
[![go vet test](https://img.shields.io/github/actions/workflow/status/YNGXZRIST/gophermart-loyalty/statictest.yml?branch=main&label=go%20vet%20test&logo=github)](https://github.com/YNGXZRIST/gophermart-loyalty/actions/workflows/statictest.yml?query=branch%3Amain)
[![coverage](https://img.shields.io/github/actions/workflow/status/YNGXZRIST/gophermart-loyalty/coverage.yml?branch=main&label=coverage&logo=github)](https://github.com/YNGXZRIST/gophermart-loyalty/actions/workflows/coverage.yml?query=branch%3Amain)
[![coverage %](.badges/coverage.svg)](https://github.com/YNGXZRIST/gophermart-loyalty/actions/workflows/coverage.yml)

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