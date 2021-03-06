require_relative '../spec_helper'

describe Dapp::Dimg::Build::Stage do
  include SpecHelper::Dimg

  after :each do
    expect(@stages).to be_empty
  end

  def dimg_last_stage
    dimg_base_last_stage(empty_dimg)
  end

  def artifact_last_stage
    dimg_base_last_stage(empty_artifact)
  end

  def dimg_base_last_stage(dimg)
    dimg.send(:last_stage_class).new(dimg)
  end

  def dimg_stages
    [:from, :before_install, :before_install_artifact, :g_a_archive, :g_a_pre_install_patch, :install, :g_a_post_install_patch,
     :after_install_artifact, :before_setup, :before_setup_artifact, :g_a_pre_setup_patch, :setup,
     :g_a_post_setup_patch, :after_setup_artifact, :g_a_latest_patch, :docker_instructions]
  end

  def artifact_stages
    [:from, :before_install, :before_install_artifact, :g_a_archive, :g_a_pre_install_patch, :install,
     :g_a_post_install_patch, :after_install_artifact, :before_setup, :before_setup_artifact, :g_a_pre_setup_patch,
     :setup, :after_setup_artifact, :g_a_artifact_patch, :build_artifact]
  end

  [:dimg, :artifact].each do |dimg|
    context dimg do
      before :each do
        @dimg = dimg
      end

      context :stages do
        before :each do
          @stages = send("#{dimg}_stages")
        end

        def first_stage
          stage = send("#{@dimg}_last_stage")
          stage = stage.prev_stage while stage.prev_stage
          stage
        end

        it 'prev_stage' do
          stage = send("#{@dimg}_last_stage")
          while stage
            expect(stage.send(:name)).to eq @stages.pop
            stage = stage.prev_stage
          end
        end

        it 'next_stage' do
          stage = first_stage
          while stage
            expect(stage.send(:name)).to eq @stages.shift
            stage = stage.next_stage
          end
        end
      end
    end
  end
end
