'use client';

import React from 'react';
import { cn } from './utils';

// =====================================
// Spinner Component
// =====================================

interface SpinnerProps {
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  className?: string;
}

const spinnerSizes = {
  xs: 'h-3 w-3',
  sm: 'h-4 w-4',
  md: 'h-6 w-6',
  lg: 'h-8 w-8',
  xl: 'h-12 w-12',
};

export function Spinner({ size = 'md', className }: SpinnerProps) {
  return (
    <svg
      className={cn('animate-spin text-current', spinnerSizes[size], className)}
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
      aria-hidden="true"
    >
      <circle
        className="opacity-25"
        cx="12"
        cy="12"
        r="10"
        stroke="currentColor"
        strokeWidth="4"
      />
      <path
        className="opacity-75"
        fill="currentColor"
        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
      />
    </svg>
  );
}

// =====================================
// Loading Overlay Component
// =====================================

interface LoadingOverlayProps {
  isLoading: boolean;
  children: React.ReactNode;
  text?: string;
  blur?: boolean;
  spinnerSize?: SpinnerProps['size'];
}

export function LoadingOverlay({
  isLoading,
  children,
  text,
  blur = true,
  spinnerSize = 'lg',
}: LoadingOverlayProps) {
  return (
    <div className="relative">
      {children}
      {isLoading && (
        <div
          className={cn(
            'absolute inset-0 z-50 flex flex-col items-center justify-center bg-white/80',
            blur && 'backdrop-blur-sm'
          )}
        >
          <Spinner size={spinnerSize} className="text-blue-600" />
          {text && <p className="mt-2 text-sm text-gray-600">{text}</p>}
        </div>
      )}
    </div>
  );
}

// =====================================
// Full Page Loading Component
// =====================================

interface PageLoadingProps {
  text?: string;
}

export function PageLoading({ text = 'Loading...' }: PageLoadingProps) {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center">
      <Spinner size="xl" className="text-blue-600" />
      <p className="mt-4 text-lg text-gray-600">{text}</p>
    </div>
  );
}

// =====================================
// Skeleton Components
// =====================================

interface SkeletonProps {
  className?: string;
  animated?: boolean;
}

export function Skeleton({ className, animated = true }: SkeletonProps) {
  return (
    <div
      className={cn(
        'rounded bg-gray-200',
        animated && 'animate-pulse',
        className
      )}
    />
  );
}

export function SkeletonText({
  lines = 3,
  className,
}: {
  lines?: number;
  className?: string;
}) {
  return (
    <div className={cn('space-y-2', className)}>
      {Array.from({ length: lines }).map((_, i) => (
        <Skeleton
          key={i}
          className={cn('h-4', i === lines - 1 ? 'w-3/4' : 'w-full')}
        />
      ))}
    </div>
  );
}

export function SkeletonCard({ className }: { className?: string }) {
  return (
    <div
      className={cn(
        'rounded-lg border border-gray-200 bg-white p-4',
        className
      )}
    >
      <div className="flex items-center space-x-4">
        <Skeleton className="h-12 w-12 rounded-full" />
        <div className="flex-1 space-y-2">
          <Skeleton className="h-4 w-1/2" />
          <Skeleton className="h-3 w-3/4" />
        </div>
      </div>
      <div className="mt-4 space-y-2">
        <Skeleton className="h-3 w-full" />
        <Skeleton className="h-3 w-full" />
        <Skeleton className="h-3 w-2/3" />
      </div>
    </div>
  );
}

export function SkeletonTable({
  rows = 5,
  columns = 4,
  className,
}: {
  rows?: number;
  columns?: number;
  className?: string;
}) {
  return (
    <div className={cn('w-full', className)}>
      {/* Header */}
      <div className="mb-2 flex space-x-4 border-b border-gray-200 pb-2">
        {Array.from({ length: columns }).map((_, i) => (
          <Skeleton key={i} className="h-4 flex-1" />
        ))}
      </div>
      {/* Rows */}
      <div className="space-y-2">
        {Array.from({ length: rows }).map((_, rowIndex) => (
          <div key={rowIndex} className="flex space-x-4 py-2">
            {Array.from({ length: columns }).map((_, colIndex) => (
              <Skeleton key={colIndex} className="h-4 flex-1" />
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}

export function SkeletonChart({ className }: { className?: string }) {
  const barHeights = [60, 80, 40, 90, 50, 70, 85, 45, 75, 55, 95, 65];
  return (
    <div className={cn('rounded-lg border border-gray-200 bg-white p-4', className)}>
      <Skeleton className="mb-4 h-4 w-1/4" />
      <div className="flex h-48 items-end space-x-2">
        {barHeights.map((height, i) => (
          <Skeleton
            key={i}
            className="flex-1"
            style={{ height: `${height}%` }}
          />
        ))}
      </div>
    </div>
  );
}

// =====================================
// Button Loading State
// =====================================

interface LoadingButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  isLoading?: boolean;
  loadingText?: string;
  spinnerSize?: SpinnerProps['size'];
}

export function LoadingButton({
  isLoading,
  loadingText,
  children,
  disabled,
  className,
  spinnerSize = 'sm',
  ...props
}: LoadingButtonProps) {
  return (
    <button
      disabled={isLoading || disabled}
      className={cn(
        'inline-flex items-center justify-center rounded-md px-4 py-2 text-sm font-medium transition-colors',
        'focus:outline-none focus:ring-2 focus:ring-offset-2',
        isLoading && 'cursor-not-allowed opacity-70',
        className
      )}
      {...props}
    >
      {isLoading && <Spinner size={spinnerSize} className="mr-2" />}
      {isLoading && loadingText ? loadingText : children}
    </button>
  );
}

// =====================================
// Content Loading States
// =====================================

interface LoadingStateProps {
  isLoading: boolean;
  isEmpty?: boolean;
  isError?: boolean;
  error?: Error | null;
  loadingComponent?: React.ReactNode;
  emptyComponent?: React.ReactNode;
  errorComponent?: React.ReactNode;
  children: React.ReactNode;
  onRetry?: () => void;
}

export function LoadingState({
  isLoading,
  isEmpty = false,
  isError = false,
  error,
  loadingComponent,
  emptyComponent,
  errorComponent,
  children,
  onRetry,
}: LoadingStateProps) {
  if (isLoading) {
    return <>{loadingComponent || <DefaultLoadingState />}</>;
  }

  if (isError) {
    return <>{errorComponent || <DefaultErrorState error={error} onRetry={onRetry} />}</>;
  }

  if (isEmpty) {
    return <>{emptyComponent || <DefaultEmptyState />}</>;
  }

  return <>{children}</>;
}

function DefaultLoadingState() {
  return (
    <div className="flex items-center justify-center py-12">
      <Spinner size="lg" className="text-blue-600" />
    </div>
  );
}

function DefaultErrorState({
  error,
  onRetry,
}: {
  error?: Error | null;
  onRetry?: () => void;
}) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <div className="mb-4 rounded-full bg-red-100 p-3">
        <svg
          className="h-6 w-6 text-red-600"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
      </div>
      <h3 className="mb-1 text-lg font-medium text-gray-900">
        Failed to load data
      </h3>
      <p className="mb-4 text-sm text-gray-500">
        {error?.message || 'An unexpected error occurred.'}
      </p>
      {onRetry && (
        <button
          onClick={onRetry}
          className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          Try again
        </button>
      )}
    </div>
  );
}

function DefaultEmptyState() {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <div className="mb-4 rounded-full bg-gray-100 p-3">
        <svg
          className="h-6 w-6 text-gray-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
          />
        </svg>
      </div>
      <h3 className="mb-1 text-lg font-medium text-gray-900">No data found</h3>
      <p className="text-sm text-gray-500">
        There is nothing here yet. Get started by adding some data.
      </p>
    </div>
  );
}

// =====================================
// Progress Indicators
// =====================================

interface ProgressBarProps {
  value: number;
  max?: number;
  showLabel?: boolean;
  size?: 'sm' | 'md' | 'lg';
  color?: 'blue' | 'green' | 'yellow' | 'red';
  className?: string;
}

const progressColors = {
  blue: 'bg-blue-600',
  green: 'bg-green-600',
  yellow: 'bg-yellow-500',
  red: 'bg-red-600',
};

const progressSizes = {
  sm: 'h-1',
  md: 'h-2',
  lg: 'h-4',
};

export function ProgressBar({
  value,
  max = 100,
  showLabel = false,
  size = 'md',
  color = 'blue',
  className,
}: ProgressBarProps) {
  const percentage = Math.min(100, Math.max(0, (value / max) * 100));

  return (
    <div className={className}>
      {showLabel && (
        <div className="mb-1 flex justify-between text-sm">
          <span className="text-gray-600">Progress</span>
          <span className="font-medium text-gray-900">{Math.round(percentage)}%</span>
        </div>
      )}
      <div
        className={cn('w-full overflow-hidden rounded-full bg-gray-200', progressSizes[size])}
        role="progressbar"
        aria-valuenow={value}
        aria-valuemin={0}
        aria-valuemax={max}
      >
        <div
          className={cn('h-full transition-all duration-300', progressColors[color])}
          style={{ width: `${percentage}%` }}
        />
      </div>
    </div>
  );
}

interface IndeterminateProgressProps {
  size?: 'sm' | 'md' | 'lg';
  color?: 'blue' | 'green' | 'yellow' | 'red';
  className?: string;
}

export function IndeterminateProgress({
  size = 'md',
  color = 'blue',
  className,
}: IndeterminateProgressProps) {
  return (
    <div
      className={cn('w-full overflow-hidden rounded-full bg-gray-200', progressSizes[size], className)}
      role="progressbar"
      aria-busy="true"
    >
      <div
        className={cn(
          'h-full w-1/3 animate-indeterminate',
          progressColors[color]
        )}
      />
    </div>
  );
}
