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
import { useCallback, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { ClipboardCheck, MessageSquare } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { CopyButton } from '@/components/copy-button'
import { PublicLayout } from '@/components/layout'
import { Turnstile } from '@/components/turnstile'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  NativeSelect,
  NativeSelectOption,
} from '@/components/ui/native-select'
import { Textarea } from '@/components/ui/textarea'
import { useStatus } from '@/hooks/use-status'
import { createPublicFeedback } from './api'
import type { CreatePublicFeedbackPayload, PublicFeedbackType } from './types'

const DEFAULT_FEEDBACK_FORM: CreatePublicFeedbackPayload = {
  feedback_type: 'complaint',
  contact_name: '',
  contact_email: '',
  contact_phone: '',
  title: '',
  content: '',
}

export function PublicFeedbackPage() {
  const { t } = useTranslation()
  const { status } = useStatus()
  const [form, setForm] =
    useState<CreatePublicFeedbackPayload>(DEFAULT_FEEDBACK_FORM)
  const [turnstileToken, setTurnstileToken] = useState('')
  const [trackingCode, setTrackingCode] = useState('')

  const turnstileEnabled = Boolean(
    status?.turnstile_check && status?.turnstile_site_key
  )
  const turnstileSiteKey = String(status?.turnstile_site_key || '')

  const submitMutation = useMutation({
    mutationFn: (payload: CreatePublicFeedbackPayload) =>
      createPublicFeedback(
        payload,
        turnstileEnabled ? turnstileToken : undefined
      ),
    onSuccess: (response) => {
      if (response.success === false) {
        toast.error(response.message || t('Failed to submit feedback'))
        return
      }
      const code = response.data?.tracking_code || ''
      setTrackingCode(code)
      setForm(DEFAULT_FEEDBACK_FORM)
      setTurnstileToken('')
      toast.success(t('Feedback submitted successfully'))
    },
  })

  const updateFormField = (
    field: keyof CreatePublicFeedbackPayload,
    value: string
  ) => {
    setForm((previous) => ({ ...previous, [field]: value }))
  }

  const verifyTurnstile = useCallback((token: string) => {
    setTurnstileToken(token)
  }, [])

  const expireTurnstile = useCallback(() => {
    setTurnstileToken('')
  }, [])

  const submitFeedback = () => {
    const payload: CreatePublicFeedbackPayload = {
      feedback_type: form.feedback_type,
      contact_name: form.contact_name.trim(),
      contact_email: form.contact_email.trim(),
      contact_phone: form.contact_phone.trim(),
      title: form.title.trim(),
      content: form.content.trim(),
    }

    if (!payload.contact_name) {
      toast.error(t('Please enter a contact name'))
      return
    }
    if (!payload.contact_email && !payload.contact_phone) {
      toast.error(t('Please enter an email or phone number'))
      return
    }
    if (!payload.title) {
      toast.error(t('Please enter a title'))
      return
    }
    if (!payload.content) {
      toast.error(t('Please enter feedback content'))
      return
    }
    if (turnstileEnabled && !turnstileToken) {
      toast.error(t('Please complete human verification'))
      return
    }

    submitMutation.mutate(payload)
  }

  return (
    <PublicLayout>
      <div className='mx-auto flex w-full max-w-3xl flex-col gap-5 py-8'>
        <Card className='overflow-hidden'>
          <CardHeader className='border-b'>
            <div className='flex items-start gap-3'>
              <div className='bg-primary/10 text-primary flex size-10 shrink-0 items-center justify-center rounded-lg'>
                <MessageSquare className='size-5' />
              </div>
              <div className='min-w-0 space-y-1'>
                <CardTitle>{t('Public complaints and feedback')}</CardTitle>
                <p className='text-muted-foreground text-sm'>
                  {t(
                    'Submit complaints, feedback, or other matters without signing in.'
                  )}
                </p>
              </div>
            </div>
          </CardHeader>
          <CardContent className='space-y-5 p-4 sm:p-6'>
            {trackingCode ? (
              <Alert className='border-emerald-200 bg-emerald-50 text-emerald-800'>
                <ClipboardCheck className='size-4' />
                <AlertTitle>{t('Submitted successfully')}</AlertTitle>
                <AlertDescription className='text-emerald-800'>
                  <div className='flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between'>
                    <span className='break-all'>
                      {t('Tracking code')}: {trackingCode}
                    </span>
                    <CopyButton
                      value={trackingCode}
                      variant='outline'
                      size='sm'
                      tooltip={t('Copy tracking code')}
                      successTooltip={t('Tracking code copied')}
                    >
                      {t('Copy')}
                    </CopyButton>
                  </div>
                </AlertDescription>
              </Alert>
            ) : null}

            <div className='grid gap-4 sm:grid-cols-2'>
              <div className='space-y-1.5'>
                <Label htmlFor='feedback-type'>{t('Type')}</Label>
                <NativeSelect
                  id='feedback-type'
                  className='w-full'
                  value={form.feedback_type}
                  onChange={(event) =>
                    updateFormField(
                      'feedback_type',
                      event.target.value as PublicFeedbackType
                    )
                  }
                >
                  <NativeSelectOption value='complaint'>
                    {t('Complaint')}
                  </NativeSelectOption>
                  <NativeSelectOption value='feedback'>
                    {t('Feedback')}
                  </NativeSelectOption>
                  <NativeSelectOption value='other'>
                    {t('Other')}
                  </NativeSelectOption>
                </NativeSelect>
              </div>
              <div className='space-y-1.5'>
                <Label htmlFor='feedback-contact-name'>
                  {t('Contact name')}
                </Label>
                <Input
                  id='feedback-contact-name'
                  value={form.contact_name}
                  onChange={(event) =>
                    updateFormField('contact_name', event.target.value)
                  }
                  placeholder={t('Enter contact name')}
                />
              </div>
              <div className='space-y-1.5'>
                <Label htmlFor='feedback-contact-email'>{t('Email')}</Label>
                <Input
                  id='feedback-contact-email'
                  type='email'
                  value={form.contact_email}
                  onChange={(event) =>
                    updateFormField('contact_email', event.target.value)
                  }
                  placeholder={t('Enter email')}
                />
              </div>
              <div className='space-y-1.5'>
                <Label htmlFor='feedback-contact-phone'>
                  {t('Phone number')}
                </Label>
                <Input
                  id='feedback-contact-phone'
                  value={form.contact_phone}
                  onChange={(event) =>
                    updateFormField('contact_phone', event.target.value)
                  }
                  placeholder={t('Enter phone number')}
                />
              </div>
              <div className='space-y-1.5 sm:col-span-2'>
                <Label htmlFor='feedback-title'>{t('Title')}</Label>
                <Input
                  id='feedback-title'
                  value={form.title}
                  onChange={(event) =>
                    updateFormField('title', event.target.value)
                  }
                  placeholder={t('Briefly describe the matter')}
                />
              </div>
              <div className='space-y-1.5 sm:col-span-2'>
                <Label htmlFor='feedback-content'>{t('Content')}</Label>
                <Textarea
                  id='feedback-content'
                  value={form.content}
                  onChange={(event) =>
                    updateFormField('content', event.target.value)
                  }
                  placeholder={t(
                    'Describe the situation, request, or suggestion in detail.'
                  )}
                  className='min-h-36'
                />
              </div>
            </div>

            {turnstileEnabled ? (
              <Turnstile
                siteKey={turnstileSiteKey}
                onVerify={verifyTurnstile}
                onExpire={expireTurnstile}
              />
            ) : null}

            <div className='flex justify-end'>
              <Button
                onClick={submitFeedback}
                disabled={submitMutation.isPending}
              >
                {submitMutation.isPending ? t('Submitting...') : t('Submit')}
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </PublicLayout>
  )
}
