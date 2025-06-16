import React, { useEffect } from 'react';
import { useAuth } from '../context/AuthContext';

type PrivateRouteProps = {
  children;
  onRequireAuth: () => void;
};

export default function PrivateRoute({ children, onRequireAuth }: PrivateRouteProps) {
  const { loggedIn } = useAuth();

  useEffect(() => {
    if (!loggedIn()) {
      onRequireAuth();
    }
  }, [loggedIn, onRequireAuth]);

  if (!loggedIn()) {
    return null; // Пока не авторизован — ничего не рендерим
  }

  return children;
}