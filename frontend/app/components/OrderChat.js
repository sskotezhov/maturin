'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { apiFetch } from 'utils/apiClient';

function formatDate(iso) {
  if (!iso) return '';
  return new Date(iso).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
}

export default function OrderChat({ orderId, currentUserId }) {
  const [messages, setMessages] = useState([]);
  const [text,     setText]     = useState('');
  const [sending,  setSending]  = useState(false);
  const [loading,  setLoading]  = useState(true);
  const bottomRef = useRef(null);

  const fetchMessages = useCallback(async () => {
    try {
      const res = await apiFetch(`/orders/${orderId}/messages`);
      if (res.ok) setMessages(await res.json());
    } finally {
      setLoading(false);
    }
  }, [orderId]);

  useEffect(() => { fetchMessages(); }, [fetchMessages]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSend = async () => {
    if (!text.trim()) return;
    setSending(true);
    try {
      const res = await apiFetch(`/orders/${orderId}/messages`, {
        method: 'POST',
        body: JSON.stringify({ text: text.trim() }),
      });
      if (res.ok) {
        setText('');
        await fetchMessages();
      }
    } finally {
      setSending(false);
    }
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleSend(); }
  };

  if (loading) return <p className="order-chat-loading">Загрузка переписки...</p>;

  return (
    <div className="order-chat">
      <div className="order-chat-messages">
        {messages.length === 0 && (
          <p className="order-chat-empty">Сообщений пока нет</p>
        )}
        {messages.map((msg) => {
          const isMine = msg.user_id === currentUserId;
          return (
            <div key={msg.id} className={`order-chat-msg ${isMine ? 'mine' : 'theirs'}`}>
              <span className="order-chat-text">{msg.text}</span>
              <span className="order-chat-time">{formatDate(msg.created_at)}</span>
            </div>
          );
        })}
        <div ref={bottomRef} />
      </div>
      <div className="order-chat-input-row">
        <textarea
          className="order-chat-input"
          placeholder="Написать сообщение..."
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={handleKeyDown}
          rows={1}
          disabled={sending}
        />
        <button
          className="order-chat-send"
          onClick={handleSend}
          disabled={sending || !text.trim()}
          aria-label="Отправить"
        >
          ➤
        </button>
      </div>
    </div>
  );
}
