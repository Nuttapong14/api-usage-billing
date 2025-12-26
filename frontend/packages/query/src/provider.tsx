'use client';

import React, { ReactNode, useState } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

interface QueryProviderProps {
  children: ReactNode;
  defaultOptions?: Parameters<typeof QueryClient>[0];
}

const defaultQueryClientOptions: Parameters<typeof QueryClient>[0] = {
  defaultOptions: {
    queries: {
      // Stale time - how long data is considered fresh
      staleTime: 1000 * 60 * 5, // 5 minutes
      // Cache time - how long inactive data is kept in cache
      gcTime: 1000 * 60 * 30, // 30 minutes
      // Retry configuration
      retry: (failureCount, error) => {
        // Don't retry on 4xx errors
        if (error && typeof error === 'object' && 'status' in error) {
          const status = (error as { status: number }).status;
          if (status >= 400 && status < 500) {
            return false;
          }
        }
        return failureCount < 3;
      },
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
      // Refetch configuration
      refetchOnWindowFocus: true,
      refetchOnReconnect: true,
      refetchOnMount: true,
    },
    mutations: {
      retry: false,
      onError: (error) => {
        console.error('Mutation error:', error);
      },
    },
  },
};

export function QueryProvider({ children, defaultOptions }: QueryProviderProps) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        ...defaultQueryClientOptions,
        ...defaultOptions,
        defaultOptions: {
          ...defaultQueryClientOptions.defaultOptions,
          ...defaultOptions?.defaultOptions,
          queries: {
            ...defaultQueryClientOptions.defaultOptions?.queries,
            ...defaultOptions?.defaultOptions?.queries,
          },
          mutations: {
            ...defaultQueryClientOptions.defaultOptions?.mutations,
            ...defaultOptions?.defaultOptions?.mutations,
          },
        },
      })
  );

  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
}

export { QueryClient, QueryClientProvider };
