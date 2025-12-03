import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import Layout from './components/Layout'
import ScrollManager from './components/ScrollManager'
import { AuthProvider } from './context/AuthContext'
import { CartProvider } from './context/CartContext'
import PrivateRoute from './routes/PrivateRoute'
import AdminRoute from './routes/AdminRoute'
import Cart from './pages/Cart'
import Home from './pages/Home'
import Login from './pages/Login'
import NotFound from './pages/NotFound'
import Shop from './pages/Shop'
import Signup from './pages/Signup'
import ProductDetail from './pages/ProductDetail'
import AdminLayout from './pages/admin/AdminLayout'
import ProductsList from './pages/admin/ProductsList'
import ProductForm from './pages/admin/ProductForm'
import UserActions from './pages/UserActions'

function App() {
  return (
    <AuthProvider>
      <CartProvider>
        <BrowserRouter>
          <ScrollManager />
          <Routes>
            <Route element={<Layout />}>
              <Route index element={<Home />} />
              <Route path="shop" element={<Shop />} />
              <Route path="productos" element={<Shop />} />
              <Route path="productos/:id" element={<ProductDetail />} />
              <Route path="login" element={<Login />} />
              <Route path="signup" element={<Signup />} />
              <Route
                path="mis-acciones"
                element={
                  <PrivateRoute>
                    <UserActions />
                  </PrivateRoute>
                }
              />
              <Route
                path="carrito"
                element={
                  <PrivateRoute>
                    <Cart />
                  </PrivateRoute>
                }
              />
              <Route path="*" element={<NotFound />} />
            </Route>
            <Route
              path="admin"
              element={
                <AdminRoute>
                  <AdminLayout />
                </AdminRoute>
              }
            >
              <Route index element={<Navigate to="productos" replace />} />
              <Route path="productos" element={<ProductsList />} />
              <Route path="productos/nuevo" element={<ProductForm />} />
              <Route path="productos/:productId/editar" element={<ProductForm />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </CartProvider>
    </AuthProvider>
  )
}

export default App
