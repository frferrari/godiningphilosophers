# Dining Philosophers
An Go implementation for the well known dining philosophers concurrency problem

This is my first Go development, so be indulgent. It is based on a Coursera assignment for the course I followed on Go programming.

The idea is that there are 5 five philosophers around a table for dinner,
Each of them have a plate, but there are only 5 chopsticks, each philosopher has 1 chopstick on his left, and one chopstick on his right.

We should implement a program that allows each philosopher to eat 3 times, knowing that philosophers who are neighborhood could not
eat at the same time because they would use the same chopstick (left for philosopher A, right for philosopher B).

Each philosopher should ask the permission to the host to eat, the host could accept or reject the request.
