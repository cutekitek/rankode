import React, { useState } from 'react';
import { useApi } from '../context/api';

export default function RegisterPage(onRegister) {
  const { auth } = useApi();

  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await auth.authRegisterPost({user: {username, email, password}});
    } catch (err) {
      console.error('Registration failed', err);
    }
  };

  return (
    <div className="p-8 space-y-6">
      <h2 className="text-3xl font-extrabold text-center text-brand">Создать аккаунт</h2>
      <form className="space-y-4" onSubmit={handleSubmit}>
        <div>
          <label className="block text-sm text-text-secondary">Имя пользователя</label>
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            className="w-full mt-1 px-3 py-2 bg-surface rounded-lg focus:outline-none focus:ring-2 focus:ring-brand-light border border-border"
            placeholder="your handle"
            required
          />
        </div>
        <div>
          <label className="block text-sm text-text-secondary">Почта</label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full mt-1 px-3 py-2 bg-surface rounded-lg focus:outline-none focus:ring-2 focus:ring-brand-light border border-border"
            placeholder="you@example.com"
            required
          />
        </div>
        <div>
          <label className="block text-sm text-text-secondary">Пароль</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full mt-1 px-3 py-2 bg-surface rounded-lg focus:outline-none focus:ring-2 focus:ring-brand-light border border-border"
            placeholder="••••••••"
            required
          />
        </div>
        <button
          type="submit"
          className="w-full py-2 font-semibold rounded-lg bg-brand hover:bg-brand-dark transition-colors"
        >
          Регистрация
        </button>
      </form>
    </div>
  );
}
