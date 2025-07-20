# Backend Developer Assessment KenTech June 2025

## Overview
[cite_start]Your primary task is to develop a Game Integration API that facilitates third-party casino games on our platform[cite: 4]. [cite_start]This new service is crucial for handling all financial transactions related to gameplay[cite: 5], with two key responsibilities:
1. [cite_start]Managing user balances through interactions with an existing, somewhat unreliable, backend service[cite: 6].
2. [cite_start]Creating and updating bets dynamically based on endpoint calls[cite: 7].

## System Context
[cite_start][Image: Diagram showing "Third party casino games" <--> "Game integration API Service to be developed" <--> "Wallet"] [cite: 8, 9, 10, 11, 12, 13]

## Wallet Service Integration
[cite_start]To streamline development, we've provisioned a mock wallet service with pre-configured users and initial balances[cite: 16]:
* [cite_start]ID: 34633089486, Currency: "USD", Balance: $5,000.00 [cite: 18]
* [cite_start]ID: 34679664254, Currency: "EUR", Balance: €9,000,000,000.00 [cite: 19]
* [cite_start]ID: 34616761765, Currency: "KES", Balance: KSh 750.50 [cite: 20]
* [cite_start]ID: 34673635133, Currency: "USD", Balance: $31,415.25 [cite: 21]

[cite_start]**Important Note:** This in-memory service does not persist with data[cite: 22]. [cite_start]All transactions and balance changes will be lost upon service restart, and balances will revert to their initial states[cite: 23].

[cite_start]You can access the Wallet Service Docker image: `docker.io/kentechsp/wallet-client` [cite: 24]
[cite_start]Use the following token to authenticate requests with the wallet service: `Wj9QhLqMUPAHSNMxeT20`[cite: 25].
[cite_start]Once the Docker image is running, you can view the API documentation via Swagger at: `http://localhost:8000/swagger/index.html`[cite: 28].

## Requirements
[cite_start]The Game Integration API needs to expose five RESTful endpoints[cite: 30]:

1.  [cite_start]**Authentication**: Authenticates a player attempting to play a game[cite: 31].
    * [cite_start]Receives: User credentials (username, password)[cite: 32].
    * [cite_start]Returns: JSON Web Token (JWT) for subsequent requests[cite: 33].

2.  [cite_start]**Player Information**: Retrieves essential player details[cite: 34].
    * [cite_start]Receives: User token (JWT)[cite: 35].
    * [cite_start]Returns: user id, balance, and currency[cite: 36].

3.  [cite_start]**Withdraw**: Processes a withdrawal from a player's balance[cite: 37]. [cite_start]Each request to this endpoint should be treated as a bet placement action[cite: 37].
    * [cite_start]Receives: User token (JWT), single transaction details including currency, amount, and provider transaction id[cite: 38].
    * [cite_start]Returns: A unique transaction ID from our system, the provider transaction id, the old balance, the new balance, and the transaction status[cite: 39].

4.  [cite_start]**Deposit**: Processes a deposit into a player's account[cite: 40]. [cite_start]This request represents a bet settlement action, and the bet's status must be determined by the transaction amount: if the amount is zero, the bet is LOST; otherwise, the bet is WON[cite: 40, 41].
    * [cite_start]Receives: User token (JWT), single transaction details including currency, amount, provider transaction id, and provider withdrawn transaction id[cite: 42].
    * [cite_start]Returns: A unique transaction ID from our system, the provider transaction id, the old balance, the new balance, and the transaction status[cite: 43, 44].

5.  [cite_start]**Cancel**: Reverts to a previously processed transaction[cite: 45].
    * [cite_start]Receives: User token (JWT), Provider transaction ID which should be rollbacked[cite: 46].
    * [cite_start]Returns: Unique transaction ID from our side, provider transaction id, old balance, new balance and transaction status[cite: 47].

## Must-Haves
[cite_start]Your solution must demonstrate proficiency in the following areas[cite: 49]:
* [cite_start]**Language**: Implemented entirely in Go[cite: 50].
* [cite_start]**Database**: Data should be stored in a database of your choice[cite: 51].
* [cite_start]**Architecture**: Follows Clean Architecture principles[cite: 55].
* [cite_start]**Code Quality**: Simple, readable, and maintainable code[cite: 56].
* [cite_start]**Security**: Includes robust authentication, authorization, and mechanisms to prevent SQL injection[cite: 57].
* [cite_start]**Version Control**: Project uploaded to a Git repository[cite: 58].
* [cite_start]**Configuration**: Utilizes environment variables for all necessary configurations[cite: 59].

## Nice-to-Haves
[cite_start]Consider incorporating the following to showcase a more comprehensive solution[cite: 61]:
* [cite_start]**ORM**: Use of an Object-Relational Mapper[cite: 63].
* [cite_start]**Error Handling**: Comprehensive and graceful error handling[cite: 64].
* [cite_start]**Logging**: Effective logging for monitoring and debugging[cite: 64].
* [cite_start]**Deployment**: Clear deployment instructions[cite: 65].
* [cite_start]**Containerization**: Docker support (e.g., docker-compose) for automated setup and execution[cite: 66].
* [cite_start]**Testing**: Unit, integration, or end-to-end tests[cite: 68].
* [cite_start]**API Documentation**: Clear and concise API documentation (e.g., OpenAPI/Swagger)[cite: 69, 70].
