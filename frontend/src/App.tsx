import { useState, ChangeEvent, KeyboardEvent, useEffect, useRef } from 'react';
import 'nes.css/css/nes.min.css';
import './App.css';
import './global.css';

// Define a type for the message object
type Message = {
  text: string;
  from: 'left' | 'right';
};

function App() {
  // Use the Message type for the messages state
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputValue, setInputValue] = useState<string>('');

  // Set latest message ref
  const endOfMessagesRef = useRef<null | HTMLDivElement>(null);

  // Snap to latest message
  useEffect(() => {
    endOfMessagesRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSendMessage = () => {
    if (inputValue.trim()) {
      // Append a new message to the messages array
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
          <div className={`nes-balloon from-${message.from}`}>
            <p>{message.text}</p>
          </div>
          <i className="nes-kirby"></i>
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