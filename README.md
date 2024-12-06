# Scalable Feed System with Messaging Queue, Redis, and Document Database  

## **Overview**  
This project implements a scalable feed system for a social media platform using the Gin framework. It incorporates a messaging queue for efficient post distribution, Redis for caching and optimization, and a document database for flexible and scalable data storage. The system is optimized to handle celebrity users with millions of followers.  

## **Features**  

### **User System**  
- Create, update, delete, and list users.  
- Follow and unfollow other users.  
- Manage celebrity status for users.  

### **Post System**  
- Create, update, delete, and list posts.  
- Like/unlike posts and manage tags.  
- Retrieve posts by specific users.  

### **Feed System**  
- Display personalized feeds in reverse-chronological order.  
- Handle celebrity fanout efficiently using Redis for caching and optimizations like batching updates or lazy loading.  



## **Technologies Used**  
- **Backend Framework**: Gin (Go framework)  
- **Messaging Queue**: RabbitMQ or Kafka for efficient post distribution  
- **Redis**: Used for caching and optimizing feed delivery and celebrity fanout  
- **Document Database**: MongoDB for storing user, post, and feed data  

## **How Redis is Used**  
- **Feed Caching**: Frequently accessed feeds are cached to reduce database load.  
- **Celebrity Fanout Optimization**: Redis is used to batch and distribute updates for users with a large number of followers.  
- **Session Management**: (Optional) Manage user sessions and rate-limiting API requests.  
