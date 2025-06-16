import React, { useState, useCallback } from 'react';
import { Button } from "../components/ui/button";
import { useApi } from '../context/api';
import { useAuth } from '../context/AuthContext';

type LoginPageProps = {
  onLogin: (token: string) => void;
};

export default function LoginPage({ onLogin }: LoginPageProps) {
  const { auth } = useApi();
  const { login } = useAuth();
  const [identifier, setIdentifier] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();
      setError(null);
      setLoading(true);

      try {
        const { token } = await auth.authLoginPost({credentials:{identifier, password}});
        login(token)
        onLogin(token!);
      } catch (err: any) {
        console.log(err)
        setError("Invalid credentials");
      } finally {
        setLoading(false);
      }
    },
    [identifier, password, auth, onLogin] 
  );

  return (
    <div className="p-8 space-y-6">
      <h2 className="text-3xl font-extrabold text-center text-brand">Войти в rankode</h2>
      <form className="space-y-4" onSubmit={handleSubmit}>
        <div>
          <label className="block text-sm">Почта или имя</label>
          <input
            type="text"
            value={identifier}
            onChange={(e) => setIdentifier(e.target.value)}
            className="w-full mt-1 px-3 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-brand-light"
            placeholder="you@example.com or username"
            required
          />
        </div>
        <div>
          <label className="block text-sm">Пароль</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full mt-1 px-3 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-brand-light"
            placeholder="••••••••"
            required
          />
        </div>

        {error && <div className="text-red-500 text-sm">{error}</div>}

        <Button
          type="submit"
          className="w-full py-2 font-semibold rounded-lg bg-brand hover:bg-brand-dark transition-colors"
          disabled={loading}
        >
          {loading ? "Вход..." : "Вход"}
        </Button>
      </form>
    </div>
  );
}
