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
import { useState } from 'react'
import { Loader2, Send } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { useCountdown } from '@/hooks/use-countdown'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { sendSelfPhoneVerification, updateUserProfile } from '../../api'

interface PhoneBindDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentPhone?: string
  username: string
  displayName?: string
  onSuccess: () => void
}

export function PhoneBindDialog({
  open,
  onOpenChange,
  currentPhone,
  username,
  displayName,
  onSuccess,
}: PhoneBindDialogProps) {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)
  const [sendingCode, setSendingCode] = useState(false)
  const [phoneNumber, setPhoneNumber] = useState('')
  const [code, setCode] = useState('')
  const {
    secondsLeft,
    isActive,
    start: startCountdown,
    reset: resetCountdown,
  } = useCountdown({
    initialSeconds: 60,
  })

  const resetForm = () => {
    setPhoneNumber('')
    setCode('')
    resetCountdown()
  }

  const handleSendCode = async () => {
    if (!phoneNumber.trim()) {
      toast.error(t('Please enter your phone number'))
      return
    }

    setSendingCode(true)
    try {
      const response = await sendSelfPhoneVerification(
        'sms_bind',
        phoneNumber.trim()
      )
      if (response.success) {
        toast.success(t('Verification code sent'))
        startCountdown()
      } else {
        toast.error(response.message || t('Failed to send verification code'))
      }
    } catch (_error) {
      toast.error(t('Failed to send verification code'))
    } finally {
      setSendingCode(false)
    }
  }

  const handleBind = async () => {
    if (!phoneNumber.trim() || !code.trim()) {
      toast.error(t('Please enter phone number and verification code'))
      return
    }

    setLoading(true)
    try {
      const response = await updateUserProfile({
        username,
        display_name: displayName || username,
        phone_number: phoneNumber.trim(),
        phone_verification_code: code.trim(),
      })

      if (response.success) {
        toast.success(t('Phone number updated successfully'))
        onOpenChange(false)
        onSuccess()
        resetForm()
      } else {
        toast.error(response.message || t('Failed to update phone number'))
      }
    } catch (_error) {
      toast.error(t('Failed to update phone number'))
    } finally {
      setLoading(false)
    }
  }

  const handleOpenChange = (nextOpen: boolean) => {
    if (loading) return
    onOpenChange(nextOpen)
    if (!nextOpen) {
      resetForm()
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>{t('Bind Phone Number')}</DialogTitle>
          <DialogDescription>
            {currentPhone
              ? t(
                  'Current phone number: {{phone}}. Enter a new phone number to change.',
                  { phone: currentPhone }
                )
              : t('Bind a phone number to your account.')}
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-4 py-4'>
          <div className='space-y-2'>
            <Label htmlFor='phone-number'>{t('Phone number')}</Label>
            <Input
              id='phone-number'
              value={phoneNumber}
              onChange={(event) => setPhoneNumber(event.target.value)}
              placeholder={t('Enter your phone number')}
              disabled={loading}
              autoComplete='tel'
            />
          </div>

          <div className='space-y-2'>
            <Label htmlFor='phone-code'>{t('Verification Code')}</Label>
            <div className='flex gap-2'>
              <Input
                id='phone-code'
                value={code}
                onChange={(event) => setCode(event.target.value)}
                placeholder={t('Enter code')}
                disabled={loading}
                autoComplete='one-time-code'
              />
              <Button
                type='button'
                variant='outline'
                onClick={handleSendCode}
                disabled={
                  loading || sendingCode || isActive || !phoneNumber.trim()
                }
                className='shrink-0 gap-1.5'
              >
                {sendingCode ? (
                  <Loader2 className='h-4 w-4 animate-spin' />
                ) : (
                  <Send className='h-4 w-4' />
                )}
                {isActive
                  ? t('Resend ({{seconds}}s)', { seconds: secondsLeft })
                  : t('Send code')}
              </Button>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button
            type='button'
            variant='outline'
            onClick={() => handleOpenChange(false)}
            disabled={loading}
          >
            {t('Cancel')}
          </Button>
          <Button
            type='button'
            onClick={handleBind}
            disabled={loading || !phoneNumber.trim() || !code.trim()}
          >
            {loading && <Loader2 className='mr-2 h-4 w-4 animate-spin' />}
            {t('Confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
