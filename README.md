# gophermart-loyalty

[![gophermart](https://img.shields.io/github/actions/workflow/status/YNGXZRIST/gophermart-loyalty/gophermart.yml?branch=iter1&label=gophermart&logo=github)](https://github.com/YNGXZRIST/gophermart-loyalty/actions/workflows/gophermart.yml?query=branch%3Aiter1)
[![go vet test](https://img.shields.io/github/actions/workflow/status/YNGXZRIST/gophermart-loyalty/statictest.yml?branch=iter1&label=go%20vet%20test&logo=github)](https://github.com/YNGXZRIST/gophermart-loyalty/actions/workflows/statictest.yml?query=branch%3Aiter1)
[![coverage](https://img.shields.io/github/actions/workflow/status/YNGXZRIST/gophermart-loyalty/coverage.yml?branch=iter1&label=coverage&logo=github)](https://github.com/YNGXZRIST/gophermart-loyalty/actions/workflows/coverage.yml?query=branch%3Aiter1)
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