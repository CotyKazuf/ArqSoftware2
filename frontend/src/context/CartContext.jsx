import { createContext, useContext, useEffect, useMemo, useReducer } from 'react'
import { createCartItem, updateCartItem } from '../services/cartService'

const CartContext = createContext(null)
const CART_STORAGE_KEY = 'lokis-perfume-cart'
const MAX_QTY = 99

const initialState = {
  items: [],
}

function cartReducer(state, action) {
  switch (action.type) {
    case 'add': {
      const { product, quantity } = action.payload
      const existing = state.items.find((item) => item.id === product.id)

      if (existing) {
        return {
          ...state,
          items: state.items.map((item) =>
            item.id === product.id
              ? { ...item, quantity: Math.min(item.quantity + quantity, MAX_QTY) }
              : item,
          ),
        }
      }

      return {
        ...state,
        items: [
          ...state.items,
          {
            ...product,
            quantity: Math.min(quantity, MAX_QTY),
          },
        ],
      }
    }
    case 'update': {
      const { id, quantity } = action.payload
      if (quantity <= 0) {
        return {
          ...state,
          items: state.items.filter((item) => item.id !== id),
        }
      }

      return {
        ...state,
        items: state.items.map((item) =>
          item.id === id ? { ...item, quantity: Math.min(quantity, MAX_QTY) } : item,
        ),
      }
    }
    case 'remove':
      return {
        ...state,
        items: state.items.filter((item) => item.id !== action.payload.id),
      }
    case 'clear':
      return initialState
    default:
      return state
  }
}

function getInitialState() {
  if (typeof window === 'undefined') return initialState

  try {
    const saved = window.localStorage.getItem(CART_STORAGE_KEY)
    if (!saved) return initialState
    const parsed = JSON.parse(saved)
    if (Array.isArray(parsed.items)) {
      return { items: parsed.items }
    }
  } catch {
    // ignore corrupted storage
  }

  return initialState
}

const resolveItemPrice = (item) => {
  const priceValue = typeof item.precio === 'number' ? item.precio : item.precioUSD
  return typeof priceValue === 'number' ? priceValue : 0
}

export function CartProvider({ children }) {
  const [state, dispatch] = useReducer(cartReducer, undefined, getInitialState)

  useEffect(() => {
    try {
      window.localStorage.setItem(CART_STORAGE_KEY, JSON.stringify(state))
    } catch {
      // ignore write errors
    }
  }, [state])

  const value = useMemo(() => {
    const totalItems = state.items.reduce((acc, item) => acc + item.quantity, 0)
    const totalPrice = state.items.reduce(
      (acc, item) => acc + resolveItemPrice(item) * item.quantity,
      0,
    )

    return {
      items: state.items,
      totalItems,
      totalPrice,
      addItem: (product, quantity = 1) => {
        dispatch({ type: 'add', payload: { product, quantity } })
        createCartItem({ ...product, quantity }).catch(() => {
          // no-op demo
        })
      },
      updateQuantity: (id, quantity) => {
        dispatch({ type: 'update', payload: { id, quantity } })
        updateCartItem({ id, quantity }).catch(() => {
          // no-op demo
        })
      },
      removeItem: (id) => dispatch({ type: 'remove', payload: { id } }),
      clearCart: () => dispatch({ type: 'clear' }),
    }
  }, [state.items])

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>
}

// eslint-disable-next-line react-refresh/only-export-components
export function useCart() {
  const context = useContext(CartContext)
  if (!context) {
    throw new Error('useCart debe usarse dentro de CartProvider')
  }
  return context
}
