import React, { createContext, useContext, useState, ReactNode } from 'react';
import Toast, { ToastType, ToastProps } from './Toast';

interface ToastContextType {
  showToast: (message: string, type?: ToastType, options?: ToastOptions) => void;
  hideToast: () => void;
}

interface ToastOptions {
  duration?: number;
  action?: {
    label: string;
    onPress: () => void;
  };
}

interface ToastState {
  visible: boolean;
  message: string;
  type: ToastType;
  duration: number;
  action?: ToastOptions['action'];
}

const ToastContext = createContext<ToastContextType | undefined>(undefined);

interface ToastProviderProps {
  children: ReactNode;
}

export const ToastProvider: React.FC<ToastProviderProps> = ({ children }) => {
  const [toastState, setToastState] = useState<ToastState>({
    visible: false,
    message: '',
    type: 'info',
    duration: 4000,
  });

  const showToast = (message: string, type: ToastType = 'info', options?: ToastOptions) => {
    setToastState({
      visible: true,
      message,
      type,
      duration: options?.duration || 4000,
      action: options?.action,
    });
  };

  const hideToast = () => {
    setToastState(prev => ({
      ...prev,
      visible: false,
    }));
  };

  return (
    <ToastContext.Provider value={{ showToast, hideToast }}>
      {children}
      <Toast
        visible={toastState.visible}
        message={toastState.message}
        type={toastState.type}
        duration={toastState.duration}
        action={toastState.action}
        onHide={hideToast}
      />
    </ToastContext.Provider>
  );
};

export const useToast = (): ToastContextType => {
  const context = useContext(ToastContext);
  if (context === undefined) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return context;
};

// Convenience methods
export const useToastHelpers = () => {
  const { showToast } = useToast();

  return {
    showSuccess: (message: string, options?: ToastOptions) => showToast(message, 'success', options),
    showError: (message: string, options?: ToastOptions) => showToast(message, 'error', options),
    showWarning: (message: string, options?: ToastOptions) => showToast(message, 'warning', options),
    showInfo: (message: string, options?: ToastOptions) => showToast(message, 'info', options),
  };
};