import { useState, ChangeEvent, KeyboardEvent, useEffect, useRef } from 'react';
import './App.css';
import './global.css';

// Define a type for the message object
type Message = {
  text: string;
  from: 'left' | 'right';
};

// bestest random ever
const senderId = Math.random().toString(36).substr(2, 9);

function App() {
  // Use the Message type for the messages state
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputValue, setInputValue] = useState<string>('');
  const ws = useRef<WebSocket | null>(null);

  // Set latest message ref
  const endOfMessagesRef = useRef<null | HTMLDivElement>(null);

  const sender = `user-${senderId}`;

  // get messages from rollup ws
  useEffect(() => {
    ws.current = new WebSocket(import.meta.env.VITE_APP_WEBSOCKET_URL);
    ws.current.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.message && data.sender !== sender) {
        const message: Message = {
          text: data.message,
          from: 'left',
        };
        setMessages((prevMessages) => [...prevMessages, message]);
      }
    };
    return () => {
      ws.current?.close();
    };
  }, [sender]);

  // Snap to latest message
  useEffect(() => {
    endOfMessagesRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSendMessage = () => {
    if (inputValue.trim()) {
      // Send to rollup api
      fetch(`${import.meta.env.VITE_APP_API_URL}/message`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          sender,
          message: inputValue
        })
      });
      // be optimistic. Append a new message to the messages array
      setMessages([...messages, { text: inputValue, from: 'right' }]);
      setInputValue('');
    }
  };

  // Add the type for the event parameter
  const handleInputChange = (event: ChangeEvent<HTMLInputElement>) => {
    setInputValue(event.target.value);
  };

  // Add the type for the event parameter
  const handleKeyPress = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter') {
      handleSendMessage();
    }
  };

  return (
    <div className="nes-container with-title is-centered">
      <div className="nes-balloon from-left">
        <p>NES.tia Chat</p>
      </div>
      <div className="message-list">
      {messages.map((message, index) => (
        <section key={index} className={`message -${message.from}`}>
          {message.from === 'left' && <i className="nes-mario"></i>}
          <div className={`nes-balloon from-${message.from}`}>
            <p>{message.text}</p>
          </div>
          {message.from === 'right' && <i className="nes-kirby"></i>}
        </section>
      ))}
      <div ref={endOfMessagesRef} />
      </div>
      <div className="nes-field is-inline">
        <input
          type="text"
          className="nes-input"
          placeholder="Type a message..."
          value={inputValue}
          onChange={handleInputChange}
          onKeyPress={handleKeyPress}
        />
        <button
          type="button"
          className="nes-btn is-primary"
          onClick={handleSendMessage}
        >
          Send
        </button>
      </div>
    </div>
  );
}

export default App;
