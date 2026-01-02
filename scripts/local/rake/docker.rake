# frozen_string_literal: true

require 'English'

namespace :docker do
  namespace :compose do

    namespace :kafka do
      desc 'run the kafka and kafka-ui only'
      task :up do
        system %{ docker compose -f docker-compose.kafka.yml up --build }
        exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
      rescue Interrupt
        exit(0)
      end

      desc 'stop the kafka and kafka-ui only'
      task :down do
        system %{ docker compose -f docker-compose.kafka.yml down --remove-orphans }
        exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
      rescue Interrupt
        exit(0)
      end
    end

    namespace :infra do
      desc 'run the infra with all components'
      task :up do
        system %{ docker compose -f docker-compose.infra.yml up --build }
        exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
      rescue Interrupt
        exit(0)
      end

      desc 'stop the infra with all components'
      task :down do
        system %{ docker compose -f docker-compose.infra.yml down --remove-orphans }
        exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
      rescue Interrupt
        exit(0)
      end
    end

  end
end
