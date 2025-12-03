import { useCallback, useState } from 'react'

export function useStatus(initial = 'idle') {
  const [status, setStatus] = useState(initial)

  const setLoading = useCallback(() => setStatus('loading'), [])
  const setSuccess = useCallback(() => setStatus('success'), [])
  const setError = useCallback(() => setStatus('error'), [])

  return {
    status,
    setStatus,
    setLoading,
    setSuccess,
    setError,
    isIdle: status === 'idle',
    isLoading: status === 'loading',
    isSuccess: status === 'success',
    isError: status === 'error',
  }
}
