import { useState, useEffect, useRef } from 'react';
import NetInfo from '@react-native-community/netinfo';
import { useAppDispatch, useAppSelector } from '../store';
import { testApiConnection } from '../store/slices/apiConfigSlice';

interface NetworkState {
  isConnected: boolean;
  type: string;
  isInternetReachable: boolean | null;
  isApiReachable: boolean;
  lastApiCheck: Date | null;
  connectionQuality: 'poor' | 'fair' | 'good' | 'excellent' | 'unknown';
}

export const useNetworkMonitor = () => {
  const dispatch = useAppDispatch();
  const { baseUrl, isConnected: apiConnected } = useAppSelector(state => state.apiConfig);

  const [networkState, setNetworkState] = useState<NetworkState>({
    isConnected: false,
    type: 'unknown',
    isInternetReachable: null,
    isApiReachable: false,
    lastApiCheck: null,
    connectionQuality: 'unknown',
  });

  const lastApiCheckRef = useRef<Date | null>(null);
  const apiCheckIntervalRef = useRef<NodeJS.Timeout | null>(null);

  // Monitor basic network connectivity
  useEffect(() => {
    const unsubscribe = NetInfo.addEventListener(state => {
      setNetworkState(prev => ({
        ...prev,
        isConnected: state.isConnected ?? false,
        type: state.type,
        isInternetReachable: state.isInternetReachable,
        connectionQuality: getConnectionQuality(state),
      }));

      // If network comes back online, test API connection
      if (state.isConnected && !apiConnected) {
        checkApiConnection();
      }
    });

    return unsubscribe;
  }, [baseUrl, apiConnected]);

  // Periodic API health check
  useEffect(() => {
    const checkInterval = 60000; // Check every minute

    const startPeriodicCheck = () => {
      if (apiCheckIntervalRef.current) {
        clearInterval(apiCheckIntervalRef.current);
      }

      apiCheckIntervalRef.current = setInterval(() => {
        if (networkState.isConnected) {
          checkApiConnection();
        }
      }, checkInterval);
    };

    startPeriodicCheck();

    return () => {
      if (apiCheckIntervalRef.current) {
        clearInterval(apiCheckIntervalRef.current);
      }
    };
  }, [networkState.isConnected, baseUrl]);

  const checkApiConnection = async () => {
    const now = new Date();

    // Avoid too frequent API checks
    if (lastApiCheckRef.current && now.getTime() - lastApiCheckRef.current.getTime() < 30000) {
      return;
    }

    lastApiCheckRef.current = now;

    try {
      const result = await dispatch(testApiConnection(baseUrl));

      setNetworkState(prev => ({
        ...prev,
        isApiReachable: testApiConnection.fulfilled.match(result),
        lastApiCheck: now,
      }));
    } catch (error) {
      setNetworkState(prev => ({
        ...prev,
        isApiReachable: false,
        lastApiCheck: now,
      }));
    }
  };

  const getConnectionQuality = (state: any): NetworkState['connectionQuality'] => {
    if (!state.isConnected) return 'poor';

    switch (state.type) {
      case 'wifi':
        return 'excellent';
      case 'cellular':
        // Could be enhanced with more detailed cellular info
        return state.details?.cellularGeneration === '4g' || state.details?.cellularGeneration === '5g'
          ? 'good'
          : 'fair';
      case 'ethernet':
        return 'excellent';
      default:
        return 'unknown';
    }
  };

  const forceApiCheck = () => {
    lastApiCheckRef.current = null;
    checkApiConnection();
  };

  const getNetworkStatusMessage = (): string => {
    if (!networkState.isConnected) {
      return '网络连接已断开';
    }

    if (!networkState.isInternetReachable) {
      return '无法访问互联网';
    }

    if (!networkState.isApiReachable) {
      return 'API服务器无法访问';
    }

    return '网络连接正常';
  };

  const getNetworkIcon = (): string => {
    if (!networkState.isConnected) return 'signal-wifi-off';

    switch (networkState.connectionQuality) {
      case 'excellent': return 'signal-wifi-4-bar';
      case 'good': return 'signal-wifi-3-bar';
      case 'fair': return 'signal-wifi-2-bar';
      case 'poor': return 'signal-wifi-1-bar';
      default: return 'signal-wifi-off';
    }
  };

  return {
    ...networkState,
    forceApiCheck,
    getNetworkStatusMessage,
    getNetworkIcon,
    isFullyConnected: networkState.isConnected && networkState.isInternetReachable && networkState.isApiReachable,
  };
};

// Hook for API retry logic
export const useApiRetry = () => {
  const { isFullyConnected } = useNetworkMonitor();
  const [retryCount, setRetryCount] = useState(0);
  const [isRetrying, setIsRetrying] = useState(false);

  const executeWithRetry = async <T>(
    apiCall: () => Promise<T>,
    maxRetries: number = 3,
    retryDelay: number = 1000
  ): Promise<T> => {
    setIsRetrying(true);

    for (let attempt = 0; attempt <= maxRetries; attempt++) {
      try {
        const result = await apiCall();
        setRetryCount(0);
        setIsRetrying(false);
        return result;
      } catch (error: any) {
        setRetryCount(attempt + 1);

        // Don't retry if network is completely down
        if (!isFullyConnected && attempt < maxRetries) {
          await new Promise(resolve => setTimeout(resolve, retryDelay * Math.pow(2, attempt)));
          continue;
        }

        // Last attempt or non-network error
        if (attempt === maxRetries) {
          setIsRetrying(false);
          throw error;
        }

        // Exponential backoff
        await new Promise(resolve => setTimeout(resolve, retryDelay * Math.pow(2, attempt)));
      }
    }

    setIsRetrying(false);
    throw new Error('Maximum retry attempts reached');
  };

  return {
    executeWithRetry,
    retryCount,
    isRetrying,
    resetRetryCount: () => setRetryCount(0),
  };
};

// Hook for offline queue
export const useOfflineQueue = () => {
  const [queue, setQueue] = useState<Array<{
    id: string;
    action: () => Promise<any>;
    description: string;
    timestamp: Date;
  }>>([]);
  const { isFullyConnected } = useNetworkMonitor();

  const addToQueue = (action: () => Promise<any>, description: string) => {
    const id = Date.now().toString();
    setQueue(prev => [...prev, {
      id,
      action,
      description,
      timestamp: new Date(),
    }]);
    return id;
  };

  const removeFromQueue = (id: string) => {
    setQueue(prev => prev.filter(item => item.id !== id));
  };

  const processQueue = async () => {
    if (!isFullyConnected || queue.length === 0) return;

    const currentQueue = [...queue];

    for (const item of currentQueue) {
      try {
        await item.action();
        removeFromQueue(item.id);
      } catch (error) {
        console.error(`Failed to process queued action: ${item.description}`, error);
        // Keep item in queue for retry later
      }
    }
  };

  // Auto-process queue when connection is restored
  useEffect(() => {
    if (isFullyConnected && queue.length > 0) {
      processQueue();
    }
  }, [isFullyConnected, queue.length]);

  return {
    queue,
    addToQueue,
    removeFromQueue,
    processQueue,
    queueSize: queue.length,
  };
};