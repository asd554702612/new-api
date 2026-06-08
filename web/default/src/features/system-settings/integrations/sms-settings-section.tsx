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
import * as z from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useTranslation } from 'react-i18next'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import {
  SettingsForm,
  SettingsSwitchContent,
  SettingsSwitchItem,
} from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { useResetForm } from '../hooks/use-reset-form'
import { useUpdateOption } from '../hooks/use-update-option'

const smsSchema = z.object({
  SMSIHuyiEnabled: z.boolean(),
  SMSIHuyiAPIID: z.string(),
  SMSIHuyiAPIKey: z.string(),
  SMSIHuyiTemplateID: z.string(),
})

type SMSFormValues = z.infer<typeof smsSchema>

type SMSSettingsSectionProps = {
  defaultValues: SMSFormValues
}

export function SMSSettingsSection({
  defaultValues,
}: SMSSettingsSectionProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()

  const form = useForm<SMSFormValues>({
    resolver: zodResolver(smsSchema),
    defaultValues,
  })

  useResetForm(form, defaultValues)

  const onSubmit = async (values: SMSFormValues) => {
    const sanitized = {
      SMSIHuyiEnabled: values.SMSIHuyiEnabled,
      SMSIHuyiAPIID: values.SMSIHuyiAPIID.trim(),
      SMSIHuyiAPIKey: values.SMSIHuyiAPIKey.trim(),
      SMSIHuyiTemplateID: values.SMSIHuyiTemplateID.trim(),
    }
    const updates: Array<{ key: string; value: string | boolean }> = []

    if (sanitized.SMSIHuyiEnabled !== defaultValues.SMSIHuyiEnabled) {
      updates.push({
        key: 'SMSIHuyiEnabled',
        value: sanitized.SMSIHuyiEnabled,
      })
    }
    if (sanitized.SMSIHuyiAPIID !== defaultValues.SMSIHuyiAPIID.trim()) {
      updates.push({ key: 'SMSIHuyiAPIID', value: sanitized.SMSIHuyiAPIID })
    }
    if (sanitized.SMSIHuyiAPIKey) {
      updates.push({ key: 'SMSIHuyiAPIKey', value: sanitized.SMSIHuyiAPIKey })
    }
    if (
      sanitized.SMSIHuyiTemplateID !==
      defaultValues.SMSIHuyiTemplateID.trim()
    ) {
      updates.push({
        key: 'SMSIHuyiTemplateID',
        value: sanitized.SMSIHuyiTemplateID,
      })
    }

    for (const update of updates) {
      await updateOption.mutateAsync(update)
    }
  }

  return (
    <SettingsSection title={t('SMS Verification')}>
      <Form {...form}>
        <SettingsForm onSubmit={form.handleSubmit(onSubmit)} autoComplete='off'>
          <SettingsPageFormActions
            onSave={form.handleSubmit(onSubmit)}
            isSaving={updateOption.isPending}
            saveLabel='Save SMS settings'
          />

          <FormField
            control={form.control}
            name='SMSIHuyiEnabled'
            render={({ field }) => (
              <SettingsSwitchItem>
                <SettingsSwitchContent>
                  <FormLabel>{t('Enable IHuyi SMS')}</FormLabel>
                  <FormDescription>
                    {t(
                      'Use IHuyi to send phone verification codes. Environment variables take priority over these settings.'
                    )}
                  </FormDescription>
                </SettingsSwitchContent>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
              </SettingsSwitchItem>
            )}
          />

          <div className='grid gap-6 md:grid-cols-2'>
            <FormField
              control={form.control}
              name='SMSIHuyiAPIID'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('IHuyi API ID')}</FormLabel>
                  <FormControl>
                    <Input
                      autoComplete='off'
                      placeholder={t('Enter IHuyi API ID')}
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='SMSIHuyiTemplateID'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('IHuyi Template ID')}</FormLabel>
                  <FormControl>
                    <Input
                      autoComplete='off'
                      placeholder='309190'
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    {t('Leave blank to use the default template ID')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          <FormField
            control={form.control}
            name='SMSIHuyiAPIKey'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('IHuyi API Key')}</FormLabel>
                <FormControl>
                  <Input
                    autoComplete='off'
                    type='password'
                    placeholder={t('Enter new key to update')}
                    {...field}
                  />
                </FormControl>
                <FormDescription>
                  {t('Leave blank to keep the existing credential')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
        </SettingsForm>
      </Form>
    </SettingsSection>
  )
}
