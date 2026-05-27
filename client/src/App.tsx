import { useAuth } from './hooks/useAuth';
import { AppShell } from './components/layout/AppShell';
import { AuthScreen } from './components/layout/AuthScreen';
import { ChatWindow } from './components/chat/ChatWindow';

function App() {
  const { isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    return <AuthScreen />;
  }

  return (
    <AppShell>
      <ChatWindow />
    </AppShell>
  );
}

export default App;
