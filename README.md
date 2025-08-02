# wheres-my-pizza
Distributed restaurant order management system. RestAPI, Postgres, RabbitMq, notifications.

#### The Order service

1. **Launch order service on specific port:**

   ```sh
   ./restaurant-system --mode=order-service --port=3000
   ```

---

#### The Kitchen service

1. **Launch a general worker:**

   ```sh
   ./restaurant-system --mode=kitchen-worker --worker-name="chef_anna" --prefetch=1
   ```

2. **Launch specialized workers:**

   ```sh
   # This worker only handles dine-in orders
   ./restaurant-system --mode=kitchen-worker --worker-name="chef_mario" --order-types="dine_in" &

   # This worker only handles delivery orders
   ./restaurant-system --mode=kitchen-worker --worker-name="chef_luigi" --order-types="delivery" &
   ```

---

#### The Tracking service

1. **Launch the Tracking service:**

   ```sh
   ./restaurant-system --mode=tracking-service --port=3002
   ```

---


#### The Notification-subscriber service

1. **Launch one or more subscribers:**

   ```sh
   # Terminal 1
   ./restaurant-system --mode=notification-subscriber

   # Terminal 2
   ./restaurant-system --mode=notification-subscriber
   ```