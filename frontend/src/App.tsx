import React, { useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import LoginPage from './pages/Login';
import RegisterPage from './pages/Register';
import MainPage from './pages/MainPage';
import { AuthProvider, useAuth } from './context/AuthContext';
import CreateTaskPage from './pages/CreateTask';
import PrivateRoute from './pages/PrivateRoute';
import TaskEditPage from './pages/TaskEditPage';
import TaskSubmissionPage from './pages/TaskSubmissionPage';
import Leaderboard from './pages/Leaderboard';


function AppLayout({ children, toggleDark, openAuth, onAuthSelect }) {
  const { logout, loggedIn } = useAuth();

  return (
    <div className="min-h-screen flex flex-col bg-bg text-text-primary transition-colors">
      <header className="p-4 bg-card-bg flex justify-between items-center">
        <Link to="/" className="text-2xl font-bold text-brand">rankode</Link>
        <nav className="space-x-4">
          <Link to="/tasks" className="hover:text-brand">Задачи</Link>
          <Link to="/leaderboard" className="hover:text-brand">Таблица лидеров</Link>
          {loggedIn() ? (
            <>
              <Link to="/task/create" className="hover:text-brand">Создать задачу</Link>
              <button
                onClick={() => logout()}
                className="px-3 py-1 bg-brand rounded-lg hover:bg-brand-dark"
              >
                Выйти
              </button>
            </>
          ) : (
            <button
              onClick={() => onAuthSelect('login')}
              className="px-3 py-1 bg-brand rounded-lg hover:bg-brand-dark"
            >
              Log In
            </button>
          )}
        </nav>
      </header>

      <main className="flex-1 p-8">{children}</main>

      <footer className="p-4 bg-card-bg text-center text-text-secondary">
        © {new Date().getFullYear()} rankode.
      </footer>
    </div>
  );
}

export default function App() {
  const [dark, setDark] = useState(true);
  const [authModal, setAuthModal] = useState<null | 'login' | 'register'>(null);

  const toggleDark = () => setDark(prev => !prev);
  const openAuth = (type: 'login' | 'register') => setAuthModal(type);
  const closeAuth = () => setAuthModal(null);
  const onLogin = () => closeAuth();

  return (
    <AuthProvider>
      <Router>
        <AppLayout toggleDark={toggleDark} openAuth={dark} onAuthSelect={openAuth}>
          <Routes>
            <Route path="/" element={<MainPage />} />
            <Route path="/task/:id" element={<TaskSubmissionPage />} />
            <Route
              path="/task/create"
              element={
                <PrivateRoute onRequireAuth={() => setAuthModal('login')}>
                  <CreateTaskPage />
                </PrivateRoute>
              }
            />
            <Route
              path="/task/:id/edit"
              element={
                <PrivateRoute onRequireAuth={() => setAuthModal('login')}>
                  <TaskEditPage />
                </PrivateRoute>
              }
            />
            <Route
              path="/leaderboard"
              element={
                <Leaderboard />
              }
            />
          </Routes>
        </AppLayout>

        {authModal === 'login' && (
          <Modal onClose={closeAuth}>
            <LoginPage onLogin={onLogin} />
            <div className="text-center text-sm mt-4">
              Don't have an account?{' '}
              <button
                onClick={() => setAuthModal('register')}
                className="text-brand-light hover:underline"
              >
                Register
              </button>
            </div>
          </Modal>
        )}

        {authModal === 'register' && (
          <Modal onClose={closeAuth}>
            <RegisterPage />
            <div className="text-center text-sm mt-4">
              Already have an account?{' '}
              <button
                onClick={() => setAuthModal('login')}
                className="text-brand-light hover:underline"
              >
                Log In
              </button>
            </div>
          </Modal>
        )}
      </Router>
    </AuthProvider>
  );
}

function Modal({ children, onClose }: { children: React.ReactNode; onClose: () => void }) {
  return (
    <div className="fixed inset-0 bg-black bg-opacity-80 flex items-center justify-center z-50">
      <div className="bg-card-bg rounded-2xl shadow-lg w-full max-w-md p-6 relative">
        <button
          onClick={onClose}
          className="absolute top-3 right-3 text-text-secondary hover:text-text-primary"
        >
          ✕
        </button>
        {children}
      </div>
    </div>
  );
}