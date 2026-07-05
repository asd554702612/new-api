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
  SettingsControlGroup,
  SettingsForm,
  SettingsSwitchContent,
  SettingsSwitchItem,
} from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { useResetForm } from '../hooks/use-reset-form'
import { useUpdateOption } from '../hooks/use-update-option'
import { removeTrailingSlash } from './utils'

const createPromptGateSchema = (t: (key: string) => string) =>
  z.object({
    'promptgate.enabled': z.boolean(),
    'promptgate.base_url': z.string().refine((value) => {
      const trimmed = value.trim()
      if (!trimmed) return true
      return /^https?:\/\//.test(trimmed)
    }, t('Provide a valid URL starting with http:// or https://')),
    'promptgate.api_key': z.string(),
    'promptgate.input_enabled': z.boolean(),
    'promptgate.output_enabled': z.boolean(),
    'promptgate.stream_output_enabled': z.boolean(),
    'promptgate.stream_fail_closed': z.boolean(),
  })

type PromptGateFormValues = z.infer<ReturnType<typeof createPromptGateSchema>>

type PromptGateSettingsSectionProps = {
  defaultValues: PromptGateFormValues
}

export function PromptGateSettingsSection({
  defaultValues,
}: PromptGateSettingsSectionProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const promptGateSchema = createPromptGateSchema(t)

  const form = useForm<PromptGateFormValues>({
    resolver: zodResolver(promptGateSchema),
    defaultValues,
  })

  useResetForm(form, defaultValues)

  const onSubmit = async (values: PromptGateFormValues) => {
    const sanitized: PromptGateFormValues = {
      ...values,
      'promptgate.base_url': removeTrailingSlash(
        values['promptgate.base_url'].trim(),
      ),
      'promptgate.api_key': values['promptgate.api_key'].trim(),
    }

    const initial: PromptGateFormValues = {
      ...defaultValues,
      'promptgate.base_url': removeTrailingSlash(
        defaultValues['promptgate.base_url'].trim(),
      ),
      'promptgate.api_key': defaultValues['promptgate.api_key'].trim(),
    }

    const updates = (
      Object.keys(sanitized) as Array<keyof PromptGateFormValues>
    )
      .filter((key) => sanitized[key] !== initial[key])
      .map((key) => ({
        key,
        value: sanitized[key],
      }))

    for (const update of updates) {
      await updateOption.mutateAsync(update)
    }
  }

  return (
    <SettingsSection title={t('PromptGate')}>
      <Form {...form}>
        <SettingsForm onSubmit={form.handleSubmit(onSubmit)} autoComplete='off'>
          <SettingsPageFormActions
            onSave={form.handleSubmit(onSubmit)}
            isSaving={updateOption.isPending}
            saveLabel="Save PromptGate settings"
          />

          <SettingsControlGroup>
            <FormField
              control={form.control}
              name='promptgate.enabled'
              render={({ field }) => (
                <SettingsSwitchItem>
                  <SettingsSwitchContent>
                    <FormLabel>{t('Enable PromptGate')}</FormLabel>
                    <FormDescription>
                      {t('Apply PromptGate moderation before returning text.')}
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
          </SettingsControlGroup>

          <FormField
            control={form.control}
            name='promptgate.base_url'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('PromptGate API Base URL')}</FormLabel>
                <FormControl>
                  <Input
                    type='url'
                    inputMode='url'
                    placeholder={t('http://127.0.0.1:8080')}
                    autoComplete='off'
                    {...field}
                    onChange={(event) => field.onChange(event.target.value)}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='promptgate.api_key'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('PromptGate API Key')}</FormLabel>
                <FormControl>
                  <Input
                    type='password'
                    placeholder={t('Enter new key to update')}
                    autoComplete='new-password'
                    {...field}
                    onChange={(event) => field.onChange(event.target.value)}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <SettingsControlGroup>
            <FormField
              control={form.control}
              name='promptgate.input_enabled'
              render={({ field }) => (
                <SettingsSwitchItem>
                  <SettingsSwitchContent>
                    <FormLabel>{t('Input moderation')}</FormLabel>
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
            <FormField
              control={form.control}
              name='promptgate.output_enabled'
              render={({ field }) => (
                <SettingsSwitchItem>
                  <SettingsSwitchContent>
                    <FormLabel>{t('Output moderation')}</FormLabel>
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
            <FormField
              control={form.control}
              name='promptgate.stream_output_enabled'
              render={({ field }) => (
                <SettingsSwitchItem>
                  <SettingsSwitchContent>
                    <FormLabel>{t('Streaming output moderation')}</FormLabel>
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
            <FormField
              control={form.control}
              name='promptgate.stream_fail_closed'
              render={({ field }) => (
                <SettingsSwitchItem>
                  <SettingsSwitchContent>
                    <FormLabel>{t('Fail closed on stream errors')}</FormLabel>
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
          </SettingsControlGroup>
        </SettingsForm>
      </Form>
    </SettingsSection>
  )
}
