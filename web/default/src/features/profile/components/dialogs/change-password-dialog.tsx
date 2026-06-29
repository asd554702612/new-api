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
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Dialog } from '@/components/dialog'
import { PasswordInput } from '@/components/password-input'
import { Input } from '@/components/ui/input'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useCountdown } from '@/hooks/use-countdown'
import { sendSelfPhoneVerification, updateUserProfile } from '../../api'

// ============================================================================
// Change Password Dialog Component
// ============================================================================

interface ChangePasswordDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  username: string
  phoneNumber?: string
}

export function ChangePasswordDialog({
  open,
  onOpenChange,
  username,
  phoneNumber,
}: ChangePasswordDialogProps) {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)
  const [sendingCode, setSendingCode] = useState(false)
  const [verificationMethod, setVerificationMethod] = useState<
    'password' | 'phone'
  >('password')
  const [formData, setFormData] = useState({
    originalPassword: '',
    phoneVerificationCode: '',
    newPassword: '',
    confirmPassword: '',
  })
  const {
    secondsLeft,
    isActive,
    start: startCountdown,
    reset: resetCountdown,
  } = useCountdown({
    initialSeconds: 60,
  })

  const handleChange = (field: string, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }))
  }

  const resetForm = () => {
    setVerificationMethod('password')
    setFormData({
      originalPassword: '',
      phoneVerificationCode: '',
      newPassword: '',
      confirmPassword: '',
    })
    resetCountdown()
  }

  const handleSendPhoneCode = async () => {
    if (!phoneNumber) {
      toast.error(t('Please bind a phone number first'))
      return
    }

    setSendingCode(true)
    try {
      const response = await sendSelfPhoneVerification('sms_change_password')
      if (response.success) {
        toast.success(t('Verification code sent'))
        startCountdown()
      } else {
        toast.error(response.message || t('Failed to send verification code'))
      }
    } catch {
      toast.error(t('Failed to send verification code'))
    } finally {
      setSendingCode(false)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    // Validation
    if (verificationMethod === 'password' && !formData.originalPassword) {
      toast.error(t('Please enter your current password'))
      return
    }

    if (
      verificationMethod === 'phone' &&
      !formData.phoneVerificationCode.trim()
    ) {
      toast.error(t('Please enter the SMS verification code'))
      return
    }

    if (!formData.newPassword) {
      toast.error(t('Please enter a new password'))
      return
    }

    if (formData.newPassword.length < 8) {
      toast.error(t('Password must be at least 8 characters'))
      return
    }

    if (formData.originalPassword === formData.newPassword) {
      toast.error(t('New password must be different from current password'))
      return
    }

    if (formData.newPassword !== formData.confirmPassword) {
      toast.error(t('Passwords do not match'))
      return
    }

    try {
      setLoading(true)
      const response = await updateUserProfile({
        original_password:
          verificationMethod === 'password'
            ? formData.originalPassword
            : undefined,
        phone_verification_code:
          verificationMethod === 'phone'
            ? formData.phoneVerificationCode.trim()
            : undefined,
        password: formData.newPassword,
      })

      if (response.success) {
        toast.success(t('Password changed successfully'))
        onOpenChange(false)
        resetForm()
      } else {
        toast.error(response.message || t('Failed to change password'))
      }
    } catch {
      toast.error(t('Failed to change password'))
    } finally {
      setLoading(false)
    }
  }

  const formId = 'change-password-form'

  return (
    <Dialog
      open={open}
      onOpenChange={(nextOpen) => {
        onOpenChange(nextOpen)
        if (!nextOpen) resetForm()
      }}
      title={t('Change Password')}
      description={
        <>
          {t('Update your password for account:')} <strong>{username}</strong>
        </>
      }
      contentClassName='sm:max-w-md'
      contentHeight='auto'
      bodyClassName='space-y-4'
      footer={
        <>
          <Button
            type='button'
            variant='outline'
            onClick={() => onOpenChange(false)}
            disabled={loading}
          >
            {t('Cancel')}
          </Button>
          <Button type='submit' form={formId} disabled={loading}>
            {loading && <Loader2 className='mr-2 h-4 w-4 animate-spin' />}
            {loading ? t('Changing...') : t('Change Password')}
          </Button>
        </>
      }
    >
      <form id={formId} onSubmit={handleSubmit} className='space-y-4'>
        <Tabs
          value={verificationMethod}
          onValueChange={(value) =>
            setVerificationMethod(value as 'password' | 'phone')
          }
        >
          <TabsList className='grid w-full grid-cols-2'>
            <TabsTrigger value='password'>{t('Current password')}</TabsTrigger>
            <TabsTrigger value='phone' disabled={!phoneNumber}>
              {t('Phone code')}
            </TabsTrigger>
          </TabsList>
        </Tabs>

        <div className='space-y-2'>
          {verificationMethod === 'password' ? (
            <>
              <Label htmlFor='currentPassword'>{t('Current Password')}</Label>
              <PasswordInput
                id='currentPassword'
                value={formData.originalPassword}
                onChange={(e) =>
                  handleChange('originalPassword', e.target.value)
                }
                disabled={loading}
                required
                autoComplete='current-password'
              />
            </>
          ) : (
            <>
              <Label htmlFor='phoneVerificationCode'>
                {t('SMS verification code')}
              </Label>
              <div className='flex gap-2'>
                <Input
                  id='phoneVerificationCode'
                  value={formData.phoneVerificationCode}
                  onChange={(e) =>
                    handleChange('phoneVerificationCode', e.target.value)
                  }
                  disabled={loading}
                  autoComplete='one-time-code'
                  placeholder={t('Enter code')}
                />
                <Button
                  type='button'
                  variant='outline'
                  className='shrink-0 gap-1.5'
                  onClick={handleSendPhoneCode}
                  disabled={loading || sendingCode || isActive || !phoneNumber}
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
            </>
          )}
        </div>

        <div className='space-y-2'>
          <Label htmlFor='newPassword'>{t('New Password')}</Label>
          <PasswordInput
            id='newPassword'
            value={formData.newPassword}
            onChange={(e) => handleChange('newPassword', e.target.value)}
            disabled={loading}
            required
            minLength={8}
            autoComplete='new-password'
          />
          <p className='text-muted-foreground text-xs'>
            {t('Must be at least 8 characters')}
          </p>
        </div>

        <div className='space-y-2'>
          <Label htmlFor='confirmPassword'>{t('Confirm New Password')}</Label>
          <PasswordInput
            id='confirmPassword'
            value={formData.confirmPassword}
            onChange={(e) => handleChange('confirmPassword', e.target.value)}
            disabled={loading}
            required
            autoComplete='new-password'
          />
        </div>
      </form>
    </Dialog>
  )
}
