# Sequence Diagrams

- Cauldron HTTP Server
- GitHub Consumer

---

## Cauldron HTTP Server, Webhook Handler

```mermaid
sequenceDiagram
    participant Client
    participant HTTP_Server
    participant Kafka_Producer

    Client->>HTTP_Server: HTTP Request
    alt Path Not Known
        HTTP_Server-->>Client: 404 Not Found
    else Method Not Allowed
        HTTP_Server-->>Client: 405 Method Not Allowed
    else Path = "/v1/webhook/github" and Method = POST
        HTTP_Server->>HTTP_Server: Check HTTP Headers
        alt Missing X-Github-Delivery
            HTTP_Server-->>Client: 400 Bad Request
        else
            HTTP_Server->>HTTP_Server: Parse Payload and Validate Signature
            alt Payload Parsing or Signature Validation Error
                HTTP_Server-->>Client: 400 Bad Request
            else
                HTTP_Server->>HTTP_Server: Marshal Parsed Struct to JSON
                alt JSON Marshaling Error
                    HTTP_Server-->>Client: 500 Internal Server Error
                else
                    HTTP_Server->>Kafka_Producer: Create Kafka Message (Buffered Channel)
                    Kafka_Producer-->>HTTP_Server: Message Acknowledged
                    HTTP_Server-->>Client: 202 Accepted
                end
            end
        end
    end
```

---

## GitHub Consumer

```mermaid
sequenceDiagram
    participant Kafka_Consumer
    participant Kafka_Message
    participant DB_Storage

    Kafka_Consumer->>Kafka_Message: Consume Message
    Kafka_Message->>Kafka_Consumer: Kafka Message with Headers and Payload

    Kafka_Consumer->>Kafka_Consumer: Parse deliveryID (UUID)
    alt deliveryID Parsing Error
        Kafka_Consumer-->>Kafka_Message: Return Error
    else
        Kafka_Consumer->>Kafka_Consumer: Parse targetID and hookID (Uint64)
        alt Parsing Error
            Kafka_Consumer-->>Kafka_Message: Return Error
        else
            Kafka_Consumer->>Kafka_Consumer: Extract target, event, offset, and partition
            Kafka_Consumer->>Kafka_Consumer: Unmarshal Payload based on Event
            alt Unmarshalling Error
                Kafka_Consumer-->>Kafka_Message: Return Error
            else
                Kafka_Consumer->>Kafka_Consumer: Extract userID and userLogin from Payload
                Kafka_Consumer->>DB_Storage: Store GitHubWebhookData
                alt Storage Error
                    Kafka_Consumer-->>Kafka_Message: Return Error
                else
                    Kafka_Consumer-->>Kafka_Message: Acknowledge Message
                end
            end
        end
    end
```