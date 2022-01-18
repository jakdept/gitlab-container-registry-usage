def container_registry_storage_usage(group_name)

  group_info = Client::GitlabApp.get("/api/v4/groups/#{group_name}", token: Rails.application.secrets.gitlab_admin_api_token)

  return 0 unless group_info['id']
  group_id = group_info['id']
  group_query = Client::GitlabApp.get("/api/v4/groups/#{group_id}/registry/repositories?tags=1", token: Rails.application.secrets.gitlab_admin_api_token)
  total_pages = group_query.headers['x-total-pages'].to_i
  current_page = 1
  total_size_per_project = {}

  total_storage = 0

  for i in 1..total_pages

    gq = Client::GitlabApp.get("/api/v4/groups/#{group_name}/registry/repositories?tags=1&per_page=100&page=#{current_page}", token: Rails.application.secrets.gitlab_admin_api_token)

    gq.each do |repo|
      repo_id = repo['id']
      project_id = repo['project_id']

      project_info = Client::GitlabApp.get("/api/v4/projects/#{project_id}", token: Rails.application.secrets.gitlab_admin_api_token)
      project_name = project_info['name']
      project_path = project_info['path_with_namespace']
      total_size_per_project[project_name] = {}
      total_size_per_project[project_name]['project_id'] = project_id
      total_size_per_project[project_name]['project_path'] = project_path
      total_size_per_project[project_name]['container_registry_storage_used'] = 0


      repo['tags'].each do |t|
        tag_name =  t['name']
        tag = Client::GitlabApp.get("/api/v4/projects/#{project_id}/registry/repositories/#{repo_id}/tags/#{tag_name}", token: Rails.application.secrets.gitlab_admin_api_token)
        total_size_per_project[project_name]['container_registry_storage_used'] = total_size_per_project[project_name]['container_registry_storage_used'] + tag['total_size'].to_f
      end
      total_size_per_project[project_name]['container_registry_storage_used'] = total_size_per_project[project_name]['container_registry_storage_used'] / (1024 * 1024 * 1024)
      total_storage = total_storage + total_size_per_project[project_name]['container_registry_storage_used']
      total_size_per_project[project_name]['container_registry_storage_used'] = total_size_per_project[project_name]['container_registry_storage_used'].round(3).to_s + " GB"

    end
    current_page = current_page + 1
  end
  print "[*] Total Storage Used [ %s GB ]  \n\n" % total_storage.round(3).to_s
  pp total_size_per_project
end
