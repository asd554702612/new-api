/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useState, useCallback } from 'react'
import i18next from 'i18next'
import { toast } from 'sonner'
import {
  calculateAmount,
  calculateAlipayAmount,
  calculateStripeAmount,
  calculateWechatPayAmount,
  calculateWaffoPancakeAmount,
  requestPayment,
  requestAlipayPayment,
  requestStripePayment,
  requestWechatPayPayment,
  isApiSuccess,
} from '../api'
import {
  getDefaultOfficialTradeType,
  isAlipayDirectPayment,
  isStripePayment,
  isWechatPayPayment,
  isWaffoPancakePayment,
  submitPaymentForm,
} from '../lib'
import type { OfficialPaymentData } from '../types'

// ============================================================================
// Payment Hook
// ============================================================================

export function usePayment() {
  const [amount, setAmount] = useState<number>(0)
  const [calculating, setCalculating] = useState(false)
  const [processing, setProcessing] = useState(false)
  const [officialPayment, setOfficialPayment] =
    useState<OfficialPaymentData | null>(null)

  // Calculate payment amount
  const calculatePaymentAmount = useCallback(
    async (topupAmount: number, paymentType: string) => {
      try {
        setCalculating(true)

        const isStripe = isStripePayment(paymentType)
        const isPancake = isWaffoPancakePayment(paymentType)
        const isWechatPay = isWechatPayPayment(paymentType)
        const isAlipayDirect = isAlipayDirectPayment(paymentType)
        const response = isStripe
          ? await calculateStripeAmount({ amount: topupAmount })
          : isWechatPay
            ? await calculateWechatPayAmount({ amount: topupAmount })
            : isAlipayDirect
              ? await calculateAlipayAmount({ amount: topupAmount })
              : isPancake
                ? await calculateWaffoPancakeAmount({ amount: topupAmount })
                : await calculateAmount({ amount: topupAmount })

        if (isApiSuccess(response) && response.data) {
          const calculatedAmount = parseFloat(response.data)
          setAmount(calculatedAmount)
          return calculatedAmount
        }

        // Don't show error for calculation, just set to 0
        setAmount(0)
        return 0
      } catch (_error) {
        setAmount(0)
        return 0
      } finally {
        setCalculating(false)
      }
    },
    []
  )

  // Process payment
  const processPayment = useCallback(
    async (topupAmount: number, paymentType: string) => {
      try {
        setProcessing(true)

        const isStripe = isStripePayment(paymentType)
        const isWechatPay = isWechatPayPayment(paymentType)
        const isAlipayDirect = isAlipayDirectPayment(paymentType)
        const amount = Math.floor(topupAmount)

        const response = isStripe
          ? await requestStripePayment({
              amount,
              payment_method: 'stripe',
            })
          : isWechatPay
            ? await requestWechatPayPayment({
                amount,
                payment_method: paymentType,
                trade_type: getDefaultOfficialTradeType(paymentType),
              })
            : isAlipayDirect
              ? await requestAlipayPayment({
                  amount,
                  payment_method: paymentType,
                  trade_type: getDefaultOfficialTradeType(paymentType),
                })
              : await requestPayment({
                  amount,
                  payment_method: paymentType,
                })

        if (!isApiSuccess(response)) {
          toast.error(response.message || i18next.t('Payment request failed'))
          return false
        }

        if ((isWechatPay || isAlipayDirect) && response.data) {
          const data = response.data as OfficialPaymentData
          if (data.checkout_url) {
            window.location.href = data.checkout_url
            toast.success(i18next.t('Redirecting to payment page...'))
            return true
          }
          if (data.jsapi_params) {
            const bridge = (
              window as unknown as {
                WeixinJSBridge?: {
                  invoke: (
                    name: string,
                    params: Record<string, unknown>,
                    callback: (res: { err_msg?: string }) => void
                  ) => void
                }
              }
            ).WeixinJSBridge
            if (!bridge) {
              toast.error(i18next.t('WeChat JSAPI is not available'))
              return false
            }
            bridge.invoke(
              'getBrandWCPayRequest',
              data.jsapi_params,
              (res) => {
                if (res.err_msg === 'get_brand_wcpay_request:ok') {
                  toast.success(i18next.t('Payment initiated'))
                } else {
                  toast.error(i18next.t('Payment request failed'))
                }
              }
            )
            return true
          }
          if (data.code_url || data.qr_code) {
            setOfficialPayment(data)
            return true
          }
        }

        // Handle Stripe payment
        const stripeData = response.data as { pay_link?: string } | undefined
        if (isStripe && stripeData?.pay_link) {
          window.open(stripeData.pay_link, '_blank')
          toast.success(i18next.t('Redirecting to payment page...'))
          return true
        }

        // Handle non-Stripe payment
        if (!isStripe && response.data) {
          const url = (response as unknown as { url?: string }).url
          if (url) {
            submitPaymentForm(url, response.data as Record<string, unknown>)
            toast.success(i18next.t('Redirecting to payment page...'))
            return true
          }
        }

        return false
      } catch (_error) {
        toast.error(i18next.t('Payment request failed'))
        return false
      } finally {
        setProcessing(false)
      }
    },
    []
  )

  return {
    amount,
    calculating,
    processing,
    officialPayment,
    calculatePaymentAmount,
    processPayment,
    setAmount,
    setOfficialPayment,
  }
}
